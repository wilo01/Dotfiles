#!/usr/bin/env bash
#
# hexer-task.sh — per-task TDS Suite environments served by Hexer.
# Companion to runtds.sh (which owns the single shared dev instance).
#
# Each acorn task gets: a dedicated Oracle container cloned from a
# version-matched golden image, plus a dedicated Hexer process serving the
# task's tds-suite worktree. Machine-readable progress on stdout
# ("step=<name> pct=<n> msg=<text>") so a UI can render a progress bar.
#
# Commands:
#   setup --hexer-dir <d> --tds-dir <d> [--hexer-repo <url>] [--tds-repo <url>]
#         [--branch <b>] [--suffix <s>]   one-shot new-machine bootstrap:
#                                 clone repos, pull image, deps, cert, golden
#   doctor [--gate <branch>]      environment pre-checks; --gate = blocking
#                                 subset only, exit 1 on first failure
#   build-golden <branch>         provision + migrate a DB for <branch>, then
#                                 docker-commit it as tds-golden:<tag>
#   up <slug> --branch <b> --worktree <tds-suite-worktree> --task-root <dir>
#      --db-port <n> --hexer-port <n> [--hexer-dir <dir>] [--host <hostname>]
#      --module <name:route:static:rt> [--module ...]
#   down <slug> --task-root <dir> [--keep-db]
#   status <slug> --task-root <dir>
#
set -euo pipefail

# Bundled with acorn and materialized to ~/.config/acorn/ at runtime,
# so machine paths come from the environment (acorn passes them from its
# hexer settings), never from this script's own location.
HEXER_DEFAULT="${HEXER_TASK_HEXER_DIR:-}"
TDS="${HEXER_TASK_TDS_DIR:-}"

ORACLE_IMAGE="container-registry.oracle.com/database/express:21.3.0-xe"
GOLDEN_REPO="tds-golden"
GOLDEN_BUILD_CONTAINER="tds-golden-build"
GOLDEN_BUILD_PORT="1530"
DB_SERVICE="xepdb1"
DB_USER="coreaccess"
DB_PASS='xy*0m9'
ORACLE_PWD='TdsSuite1'

log()  { printf '\033[1;34m▶ %s\033[0m\n' "$*" >&2; }
ok()   { printf '\033[0;32m  ✔ %s\033[0m\n' "$*" >&2; }
warn() { printf '\033[0;33m  ! %s\033[0m\n' "$*" >&2; }
die()  { printf '\033[0;31m  ✘ %s\033[0m\n' "$*" >&2; exit 1; }

# Progress events consumed by acorn; everything human goes to stderr.
# When PROVISION_LOG is set (cmd_up), steps land in the log too.
step() {
   printf 'step=%s pct=%s msg=%s\n' "$1" "$2" "${3:-}"
   [ -n "${PROVISION_LOG:-}" ] && printf '▶ step %s (%s%%) %s\n' "$1" "$2" "${3:-}" >> "$PROVISION_LOG" || true
}

# `pnpm --version` (not `command -v`): the corepack shim can exist but crash
# on first use — newer Node ships corepack with stale registry signing keys
# ("Cannot find matching keyid", stack ending in "Node.js vXX"). Upgrading
# corepack refreshes the keys.
ensure_pnpm() {
   corepack enable >/dev/null 2>&1 || true
   if ! pnpm --version >/dev/null 2>&1; then
      warn "pnpm missing or its corepack shim is broken — installing corepack@latest"
      npm install -g corepack@latest >&2 2>&1 || true
      corepack enable >/dev/null 2>&1 || true
      corepack prepare pnpm@latest --activate >&2 2>&1 || true
   fi
   pnpm --version >/dev/null 2>&1 \
      || die "pnpm not working on node $(node -v) — run: npm install -g corepack@latest && corepack enable && corepack prepare pnpm@latest --activate"
}

# Runs a heredoc-fed Node script from a temp file. `node /dev/stdin` breaks
# on Linux when stdin is a pipe: Node realpaths the script entry to
# '/proc/<pid>/fd/pipe:[...]' and ENOENTs.
run_node_script() {
   local tmp rc=0
   tmp="$(mktemp)"
   cat > "$tmp"
   node "$tmp" "$@" || rc=$?
   rm -f "$tmp"
   return $rc
}

golden_tag() { echo "$1" | tr '/' '-'; }
container_name() { echo "tds-task-$1"; }

container_exists()  { docker ps -a --format '{{.Names}}' | grep -qx "$1"; }
container_running() { docker ps --format '{{.Names}}' | grep -qx "$1"; }

wait_db_ready() {
   local name="$1" tries="${2:-240}"
   local i health
   for i in $(seq 1 "$tries"); do
      health="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{end}}' "$name" 2>/dev/null || true)"
      [ "$health" = "healthy" ] && return 0
      # grep without -q: it must consume the whole stream — early exit sends
      # docker logs a SIGPIPE, which pipefail turns into a false negative.
      if docker logs "$name" 2>&1 | grep "DATABASE IS READY TO USE" >/dev/null; then
         return 0
      fi
      container_running "$name" || die "container $name stopped — check 'docker logs $name'"
      sleep 5
   done
   die "$name did not become ready — check 'docker logs $name' (Rosetta enabled?)"
}

CERT_DIR="$HOME/.config/acorn/certs"

# Self-signed wildcard cert per hostname suffix (*.acrid.dev). The .dev TLD
# is HSTS-preloaded — browsers force HTTPS, plain HTTP can never work.
ensure_dev_cert() {
   local suffix="$1"
   CERT_FILE="$CERT_DIR/$suffix.crt"
   KEY_FILE="$CERT_DIR/$suffix.key"
   [ -f "$CERT_FILE" ] && [ -f "$KEY_FILE" ] && return 0
   mkdir -p "$CERT_DIR"
   openssl req -x509 -newkey rsa:2048 -sha256 -days 3650 -nodes \
      -keyout "$KEY_FILE" -out "$CERT_FILE" \
      -subj "/CN=*.$suffix" \
      -addext "subjectAltName=DNS:*.$suffix,DNS:$suffix,DNS:localhost,IP:127.0.0.1" \
      >/dev/null 2>&1 || die "openssl cert generation failed for *.$suffix"
   chmod 600 "$KEY_FILE"
   ok "self-signed wildcard cert created for *.$suffix ($CERT_DIR)"
}

# HSTS-preloaded TLDs (.dev) hard-block untrusted certs — no "Proceed
# anyway" in Chrome. The cert must be a trusted anchor in the System
# keychain. Deliberately NOT automated: changing the system trust store is
# the developer's own call, so we detect and hand them the exact command.
cert_trust_command() {
   echo "sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain '$CERT_FILE'"
}

cert_is_trusted() {
   [ "$(uname)" = "Darwin" ] || return 0
   security verify-cert -c "$CERT_FILE" -p ssl >/dev/null 2>&1
}

ensure_cert_trusted() {
   cert_is_trusted && return 0
   die "dev certificate is not trusted — browsers hard-block .dev domains. Run once: $(cert_trust_command)"
}

# Protocol-agnostic TLS terminator: decrypts on the public task port and
# pipes raw bytes to hexer's internal HTTP port (WebSockets included).
start_tls_proxy() {
   local https_port="$1" target_port="$2" state="$3"
   nohup node -e '
const tls=require("tls"),net=require("net"),fs=require("fs");
const [cert,key,httpsPort,targetPort]=process.argv.slice(1);
const srv=tls.createServer({cert:fs.readFileSync(cert),key:fs.readFileSync(key),minVersion:"TLSv1.2"},(s)=>{
  const u=net.connect(Number(targetPort),"127.0.0.1");
  s.pipe(u).pipe(s);
  s.on("error",()=>u.destroy());u.on("error",()=>s.destroy());
});
srv.on("tlsClientError",()=>{});
srv.listen(Number(httpsPort),"0.0.0.0",()=>console.log(`tls ${httpsPort} -> http ${targetPort}`));
' "$CERT_FILE" "$KEY_FILE" "$https_port" "$target_port" \
      >"$state/logs/tls.log" 2>&1 &
   echo $! > "$state/tls.pid"
}

# Task hostnames (<slug>.acrid.dev) must resolve to loopback. On macOS a
# native admin-password dialog adds the entry; elsewhere we print the line.
ensure_host_mapping() {
   local host="$1"
   grep -qE "^[^#]*[[:space:]]$host([[:space:]]|\$)" /etc/hosts 2>/dev/null && return 0
   warn "$host missing from /etc/hosts — requesting admin approval to add it"
   if [ "$(uname)" = "Darwin" ]; then
      osascript -e "do shell script \"printf '127.0.0.1 $host\\n' >> /etc/hosts\" with administrator privileges" >/dev/null 2>&1 \
         || die "cannot add $host to /etc/hosts — run: sudo sh -c 'echo 127.0.0.1 $host >> /etc/hosts'"
      ok "$host → 127.0.0.1 added to /etc/hosts"
   else
      die "add to /etc/hosts first: sudo sh -c 'echo 127.0.0.1 $host >> /etc/hosts'"
   fi
}

kill_port_listeners() {
   local port="$1"
   local pids
   pids="$(lsof -ti "tcp:$port" 2>/dev/null || true)"
   [ -n "$pids" ] || return 0
   warn "killing stale listener(s) on port $port: $pids"
   echo "$pids" | while read -r p; do kill "$p" 2>/dev/null || true; done
   sleep 1
}

# ── JDK (same constraint as runtds.sh: Liquibase breaks on a JAVA_HOME with
# spaces, e.g. Android Studio's bundled JDK) ─────────────────────────────────
detect_java_home() {
   local jh=""
   jh="$(/usr/libexec/java_home 2>/dev/null || true)"
   if [ -z "$jh" ] || [ ! -x "$jh/bin/java" ]; then
      local asj="/Applications/Android Studio.app/Contents/jbr/Contents/Home"
      [ -x "$asj/bin/java" ] && jh="$asj"
   fi
   if { [ -z "$jh" ] || [ ! -x "$jh/bin/java" ]; } && command -v java >/dev/null; then
      jh="$(dirname "$(dirname "$(readlink -f "$(command -v java)")")")"
   fi
   [ -n "$jh" ] && [ -x "$jh/bin/java" ] && echo "$jh"
}

LB_JAVA_HOME=""
ensure_jdk() {
   local jh="${JAVA_HOME:-}"
   [ -n "$jh" ] || jh="$(detect_java_home)"
   [ -n "$jh" ] && [ -x "$jh/bin/java" ] || die "No JDK found — set JAVA_HOME to a JDK 17+ (Linux: apt install openjdk-17-jdk)"
   case "$jh" in
      *" "*) ln -sfn "$jh" "$HOME/.tds-jdk"; LB_JAVA_HOME="$HOME/.tds-jdk" ;;
      *)     LB_JAVA_HOME="$jh" ;;
   esac
}

# Self-heal the SUITE-4048 corrupt package header if the branch still has it.
patch_known_source_bug() {
   local f="$1/source/server/database/sql/safe/packages/ca_visitor_status_pak.sql"
   if grep -q "package ca_visitor_status_pak as ca_visitor_status_pak.C_CHECKED_IN" "$f" 2>/dev/null; then
      warn "patching corrupt header in ca_visitor_status_pak.sql (SUITE-4048)"
      perl -0pi -e 's/package ca_visitor_status_pak as ca_visitor_status_pak\.C_CHECKED_IN/package ca_visitor_status_pak as/' "$f"
   fi
}

# Substitute {{ssm-*}} placeholders portably (node, not GNU sed), run
# Liquibase against localhost:<port>, restore the tracked files.
run_liquibase() {
   local worktree="$1" port="$2"
   ensure_jdk
   patch_known_source_bug "$worktree"
   local rev; rev="$(cd "$worktree" && git rev-parse HEAD)"
   ( cd "$worktree"
     run_node_script "$rev" "$DB_USER" "$DB_PASS" "$port" "$DB_SERVICE" <<'NODE'
const fs = require("fs");
const [rev, user, pass, port, service] = process.argv.slice(2);
const props = "source/server/database/liquibase.properties";
const cl = "source/server/database/sql/safe/safe.changelog.xml";
let p = fs.readFileSync(props, "utf8");
p = p.replace("{{ssm-db-hostname}}:1521:{{ssm-db-sid}}", `localhost:${port}/${service}`)
     .replace("{{ssm-db-username}}", user)
     .replace("{{ssm-db-password}}", pass);
fs.writeFileSync(props, p);
let c = fs.readFileSync(cl, "utf8");
const map = {
  "{{ssm-app-version}}": "master", "{{ssm-app-revision}}": rev,
  "{{ssm-instance-name}}": "Test", "{{ssm-product-install-string}}": "(A)(C)(E)(G)(CL)(AV)",
  "{{ssm-system-type}}": "UAT", "{{ssm-default-department}}": "1", "{{ssm-default-zone}}": "101001",
  "{{ssm-allow-send-email}}": "Y", "{{ssm-smtp-auth-login}}": "Y",
  "{{ssm-smtp-default-sender-email}}": "noreply@tdscloud.ie", "{{ssm-smtp-domain-name}}": "tdscloud.ie",
  "{{ssm-smtp-port}}": "587", "{{ssm-smtp-username}}": "smtp_user", "{{ssm-smtp-password}}": "smtp_password",
  "{{ssm-smtp-wallet-password}}": "wallet_password", "{{ssm-smtp-wallet-path}}": "file://rdsdbdata/datapump",
  "{{ssm-smtp-use-static-sender-email}}": "Y", "{{ssm-smtp-use-tls}}": "Y",
};
for (const [k, v] of Object.entries(map)) c = c.split(k).join(v);
fs.writeFileSync(cl, c);
NODE
     ( cd source/server/database && rm -f install-update.log && JAVA_HOME="$LB_JAVA_HOME" ./liquibase update ) || true
     git restore source/server/database/liquibase.properties source/server/database/sql/safe/safe.changelog.xml 2>/dev/null || true
   )
}

create_db_user() {
   local name="$1"
   docker exec -i "$name" sqlplus -s "/ as sysdba" >/dev/null 2>&1 <<SQL || true
ALTER SESSION SET CONTAINER=XEPDB1;
CREATE USER $DB_USER IDENTIFIED BY "$DB_PASS";
GRANT CONNECT, RESOURCE, DBA TO $DB_USER;
GRANT SELECT ON sys.user\$ TO $DB_USER;
GRANT EXECUTE ON sys.dbms_crypto TO PUBLIC;
GRANT MANAGE SCHEDULER TO PUBLIC;
GRANT SELECT ON "GV_\$SESSION" TO $DB_USER;
EXIT;
SQL
}

recompile_schema() {
   docker exec "$1" bash -lc "echo \"begin sys.dbms_utility.compile_schema('$DB_USER',false); end;\" | sqlplus -s '/ as sysdba'" >/dev/null 2>&1 || true
}

# ── doctor ────────────────────────────────────────────────────────────────────
check() {
   # check <gate?> <name> <ok?> <detail>
   local gate="$1" name="$2" pass="$3" detail="$4"
   if [ "$pass" = "1" ]; then
      echo "ok $name: $detail"
   else
      echo "fail $name: $detail"
      DOCTOR_FAILED=1
      if [ "$GATE_MODE" = "1" ] && [ "$gate" = "1" ]; then
         die "$name — $detail"
      fi
   fi
}

cmd_doctor() {
   GATE_MODE=0
   local gate_branch=""
   while [ $# -gt 0 ]; do
      case "$1" in
         --gate)      GATE_MODE=1; gate_branch="${2:?--gate needs a branch}"; shift 2 ;;
         --hexer-dir) HEXER_DEFAULT="$2"; shift 2 ;;
         --tds-dir)   TDS="$2"; shift 2 ;;
         *) die "unknown doctor option: $1" ;;
      esac
   done
   DOCTOR_FAILED=0

   if command -v node >/dev/null; then
      local major; major="$(node -p 'process.versions.node.split(".")[0]')"
      check 1 "node" "$([ "$major" -ge 22 ] && echo 1 || echo 0)" "node $(node -v) (need >=22)"
   else
      check 1 "node" 0 "node not installed (need >=22)"
   fi
   local pnpm_version
   if pnpm_version="$(pnpm --version 2>/dev/null)"; then
      check 1 "pnpm" 1 "pnpm $pnpm_version"
   else
      check 1 "pnpm" 0 "pnpm missing or corepack shim broken — run: npm install -g corepack@latest && corepack enable"
   fi
   command -v docker >/dev/null && docker info >/dev/null 2>&1 \
      && check 1 "docker" 1 "daemon running" \
      || check 1 "docker" 0 "docker missing or daemon not running"

   if docker image inspect "$ORACLE_IMAGE" >/dev/null 2>&1; then
      check 0 "oracle-image" 1 "$ORACLE_IMAGE pulled"
   else
      check 0 "oracle-image" 0 "$ORACLE_IMAGE not pulled (~2.5 GB): docker pull $ORACLE_IMAGE"
   fi

   local jh="${JAVA_HOME:-$(detect_java_home)}"
   if [ -n "$jh" ] && [ -x "$jh/bin/java" ]; then
      case "$jh" in
         *" "*) check 0 "jdk" 1 "JDK at '$jh' (has spaces — ~/.tds-jdk symlink will be used)" ;;
         *)     check 0 "jdk" 1 "JDK at $jh" ;;
      esac
   else
      check 0 "jdk" 0 "no JDK 17+ found — Liquibase needs one (set JAVA_HOME)"
   fi

   if [ -n "$HEXER_DEFAULT" ] && [ -d "$HEXER_DEFAULT" ]; then
      check 1 "hexer-dir" 1 "$HEXER_DEFAULT"
      [ -d "$HEXER_DEFAULT/node_modules" ] && check 1 "hexer-deps" 1 "node_modules present" \
         || check 1 "hexer-deps" 0 "run: cd $HEXER_DEFAULT && pnpm install"
   else
      check 1 "hexer-dir" 0 "hexer checkout not configured or missing (set 'hexer dir' in SETTINGS)"
   fi
   if [ -n "$TDS" ] && [ -d "$TDS" ]; then
      check 1 "tds-suite" 1 "$TDS"
   else
      check 1 "tds-suite" 0 "tds-suite checkout not configured or missing (set 'tds dir' in SETTINGS)"
   fi

   if [ -n "$gate_branch" ]; then
      local tag; tag="$(golden_tag "$gate_branch")"
      docker image inspect "$GOLDEN_REPO:$tag" >/dev/null 2>&1 \
         && check 1 "golden-image" 1 "$GOLDEN_REPO:$tag" \
         || check 1 "golden-image" 0 "no golden DB for '$gate_branch' — run: $0 build-golden $gate_branch"
   else
      local images
      images="$(docker image ls --format '{{.Tag}}' "$GOLDEN_REPO" 2>/dev/null | paste -sd, - || true)"
      [ -n "$images" ] && check 0 "golden-images" 1 "$images" \
         || check 0 "golden-images" 0 "none built yet — run: $0 build-golden master"
   fi

   CERT_FILE="$CERT_DIR/acrid.dev.crt"
   if [ -f "$CERT_FILE" ]; then
      if cert_is_trusted; then
         check 0 "cert-trust" 1 "dev cert trusted in System keychain"
      else
         check 0 "cert-trust" 0 "run once: $(cert_trust_command)"
      fi
   else
      check 0 "cert-trust" 1 "dev cert not generated yet (created on first task 'up')"
   fi

   local free_gb
   # df -g is BSD-only; GNU df (Linux) rejects it and would report 0G here.
   free_gb="$(df -Pk / 2>/dev/null | awk 'NR==2 {print int($4/1048576)}')"
   [ -n "$free_gb" ] || free_gb=0
   check 0 "disk" "$([ "${free_gb:-0}" -ge 15 ] && echo 1 || echo 0)" "${free_gb}G free on / (want >=15G)"

   [ "$DOCTOR_FAILED" = "0" ]
}

# ── build-golden ──────────────────────────────────────────────────────────────
cmd_build_golden() {
   local branch="${1:?usage: build-golden <branch>}"; shift
   while [ $# -gt 0 ]; do
      case "$1" in
         --tds-dir) TDS="$2"; shift 2 ;;
         *) die "unknown build-golden option: $1" ;;
      esac
   done
   local tag; tag="$(golden_tag "$branch")"
   local build_wt="/tmp/tds-golden-build-$tag"

   log "Building golden DB image $GOLDEN_REPO:$tag from branch $branch"
   docker info >/dev/null 2>&1 || die "Docker daemon is not running"
   [ -n "$TDS" ] && [ -d "$TDS" ] || die "tds-suite checkout not configured (HEXER_TASK_TDS_DIR)"

   ( cd "$TDS" && GIT_TERMINAL_PROMPT=0 git fetch origin "$branch" ) || die "cannot fetch origin/$branch"
   if [ -d "$build_wt" ]; then
      ( cd "$TDS" && git worktree remove --force "$build_wt" ) 2>/dev/null || rm -rf "$build_wt"
   fi
   ( cd "$TDS" && git worktree add --detach "$build_wt" "origin/$branch" ) \
      || die "cannot create build worktree for origin/$branch"

   if container_exists "$GOLDEN_BUILD_CONTAINER"; then
      docker rm -f "$GOLDEN_BUILD_CONTAINER" >/dev/null
   fi
   log "Provisioning Oracle XE on port $GOLDEN_BUILD_PORT (first boot can take up to 20 min)"
   docker run -d --name "$GOLDEN_BUILD_CONTAINER" -p "$GOLDEN_BUILD_PORT:1521" \
      -e ORACLE_PWD="$ORACLE_PWD" "$ORACLE_IMAGE" >/dev/null
   wait_db_ready "$GOLDEN_BUILD_CONTAINER"
   create_db_user "$GOLDEN_BUILD_CONTAINER"
   ok "database up, user created"

   log "Applying Liquibase schema (5–15 min)"
   run_liquibase "$build_wt" "$GOLDEN_BUILD_PORT"
   recompile_schema "$GOLDEN_BUILD_CONTAINER"
   ok "schema applied"

   log "Committing image $GOLDEN_REPO:$tag"
   docker stop "$GOLDEN_BUILD_CONTAINER" >/dev/null
   docker commit "$GOLDEN_BUILD_CONTAINER" "$GOLDEN_REPO:$tag" >/dev/null
   docker rm "$GOLDEN_BUILD_CONTAINER" >/dev/null
   ( cd "$TDS" && git worktree remove --force "$build_wt" ) 2>/dev/null || rm -rf "$build_wt"
   ok "golden image ready: $GOLDEN_REPO:$tag"
}

# ── up ────────────────────────────────────────────────────────────────────────
cmd_up() {
   local slug="${1:?usage: up <slug> ...}"; shift
   local branch="" worktree="" task_root="" db_port="" hexer_port=""
   local hexer_dir="$HEXER_DEFAULT" host="" build_golden=0
   local modules=()
   while [ $# -gt 0 ]; do
      case "$1" in
         --branch)     branch="$2"; shift 2 ;;
         --worktree)   worktree="$2"; shift 2 ;;
         --task-root)  task_root="$2"; shift 2 ;;
         --db-port)    db_port="$2"; shift 2 ;;
         --hexer-port) hexer_port="$2"; shift 2 ;;
         --hexer-dir)  hexer_dir="$2"; shift 2 ;;
         --tds-dir)    TDS="$2"; shift 2 ;;
         --host)       host="$2"; shift 2 ;;
         --module)     modules+=("$2"); shift 2 ;;
         --build-golden-if-missing) build_golden=1; shift ;;
         *) die "unknown up option: $1" ;;
      esac
   done
   [ -n "$branch" ] && [ -n "$worktree" ] && [ -n "$task_root" ] || die "up needs --branch, --worktree, --task-root"
   [ -n "$db_port" ] && [ -n "$hexer_port" ] || die "up needs --db-port and --hexer-port"
   [ ${#modules[@]} -gt 0 ] || die "up needs at least one --module name:route:static:rt"
   [ -n "$host" ] || host="$slug.acrid.dev"

   local tag; tag="$(golden_tag "$branch")"
   local name; name="$(container_name "$slug")"
   local hexer_state="$task_root/.hexer"
   mkdir -p "$hexer_state/logs"

   # Everything human-readable (incl. docker/liquibase chatter on stderr)
   # is mirrored into provision.log — the TUI's `i` key shows its tail.
   PROVISION_LOG="$hexer_state/logs/provision.log"
   : > "$PROVISION_LOG"
   exec 2> >(tee -a "$PROVISION_LOG" >&2)

   step validate 5 "checking prerequisites"
   ensure_host_mapping "$host"
   docker info >/dev/null 2>&1 || die "Docker daemon is not running"
   [ -d "$worktree" ] || die "worktree not found: $worktree"
   [ -d "$hexer_dir" ] || die "hexer dir not found: $hexer_dir"
   if ! docker image inspect "$GOLDEN_REPO:$tag" >/dev/null 2>&1; then
      if [ "$build_golden" = "1" ]; then
         step build-golden 8 "building golden DB image for $branch (one-off, 30-60 min)"
         cmd_build_golden "$branch"
      else
         die "no golden image $GOLDEN_REPO:$tag — run: $0 build-golden $branch"
      fi
   fi

   step clone-db 10 "starting database container $name"
   if container_exists "$name"; then
      container_running "$name" || docker start "$name" >/dev/null
   else
      docker run -d --name "$name" -p "$db_port:1521" "$GOLDEN_REPO:$tag" >/dev/null
   fi
   step wait-db 20 "waiting for oracle to accept connections"
   wait_db_ready "$name" 180
   step liquibase 55 "applying changeset delta from the task worktree"
   run_liquibase "$worktree" "$db_port"
   recompile_schema "$name"

   # Public task port serves HTTPS (mandatory on the HSTS-preloaded .dev
   # TLD); hexer itself listens on HTTP one thousand ports up.
   local internal_port=$((hexer_port + 1000))
   ensure_dev_cert "${host#*.}"
   ensure_cert_trusted

   step start-hexer 80 "starting hexer (https :$hexer_port → http :$internal_port)"
   local jwt_secret
   jwt_secret="$(node -e 'console.log(require("crypto").randomBytes(48).toString("hex"))')"
   local instances
   instances="$(run_node_script "$slug" "$host" "$worktree" "$DB_USER" "$DB_PASS" "$db_port" "$DB_SERVICE" "${modules[@]}" <<'NODE'
const [slug, host, worktree, user, pass, port, service, ...moduleSpecs] = process.argv.slice(2);
const modules = moduleSpecs.map((spec) => {
  const [name, route, staticRel, rtRel] = spec.split(":");
  return {
    appName: name,
    type: "web",
    version: "local",
    route: `/${route.replace(/^\//, "")}`,
    static: `${worktree}/${staticRel}`,
    resourceTemplates: `${worktree}/${rtRel}`,
    authConfig: name === "kiosk"
      ? { type: "none" }
      : { type: "oracle-db", enabled: true, secure: false,
          allowedSchemas: ["COREACCESS"], defaultRoles: ["default", "developer"] },
    enabled: true,
  };
});
console.log(JSON.stringify([{
  code: slug,
  description: `acorn task ${slug}`,
  tier: "premium",
  status: "active",
  infra: {
    multiTenancy: false, environment: "development", isProduction: false,
    executionPool: 10, region: "local", primaryTimeZone: "UTC",
    sla: "business-hours", customDomain: host,
  },
  customer: "LOCAL001",
  modules,
  database: {
    user, password: pass, connectString: `localhost:${port}/${service}`,
    poolMin: 2, poolMax: 10, poolIncrement: 2,
  },
}]));
NODE
)"
   local old_pid=""
   [ -f "$hexer_state/hexer.pid" ] && old_pid="$(cat "$hexer_state/hexer.pid")"
   if [ -n "$old_pid" ] && kill -0 "$old_pid" 2>/dev/null; then
      warn "hexer already running for $slug (pid $old_pid) — restarting"
      kill "$old_pid" 2>/dev/null || true
      sleep 1
   fi
   if [ -f "$hexer_state/tls.pid" ]; then
      kill "$(cat "$hexer_state/tls.pid")" 2>/dev/null || true
      rm -f "$hexer_state/tls.pid"
   fi
   kill_port_listeners "$hexer_port"
   kill_port_listeners "$internal_port"
   # cwd = the task's state dir, NOT the hexer checkout: hexer re-reads
   # ${cwd}/.env on its periodic config reload, which would overwrite our
   # INSTANCES with the shared checkout's runtds.sh config after ~2 min.
   ( cd "$hexer_state" && \
     MODE=local NODE_ENV=development PORT="$internal_port" \
     JWT_SECRET="$jwt_secret" COOKIE_SECURE=false \
     FRONTEND_REDIRECT_URL=/safe/home \
     LOG_LEVEL=info LOG_FORMAT=simple REQUEST_LOGGING=true \
     LOGS_ROOT="$hexer_state/logs" \
     CORS_ORIGIN="$host" CORS_CREDENTIALS=true \
     CACHE_ENABLED=false CACHE_ENCRYPTION_ENABLED=false \
     LINX_USE_MOCK=false \
     INSTANCES="$instances" \
     nohup node "$hexer_dir/src/server.js" >"$hexer_state/logs/stdout.log" 2>&1 &
     echo $! > "$hexer_state/hexer.pid" )

   start_tls_proxy "$hexer_port" "$internal_port" "$hexer_state"

   step health 90 "waiting for hexer to answer"
   local i healthy=0
   for i in $(seq 1 60); do
      if curl -fsS --max-time 2 "http://localhost:$internal_port/health-check" >/dev/null 2>&1; then
         healthy=1; break
      fi
      kill -0 "$(cat "$hexer_state/hexer.pid")" 2>/dev/null \
         || die "hexer exited during startup — see $hexer_state/logs/stdout.log"
      sleep 2
   done
   [ "$healthy" = "1" ] || die "hexer did not become healthy — see $hexer_state/logs/stdout.log"
   curl -fsSk --max-time 3 "https://localhost:$hexer_port/health-check" >/dev/null 2>&1 \
      || die "TLS proxy on :$hexer_port not answering — see $hexer_state/logs/tls.log"

   step done 100 "https://$host:$hexer_port"
   ok "task environment up: https://$host:$hexer_port (db localhost:$db_port)"
}

# ── down / status ─────────────────────────────────────────────────────────────
cmd_down() {
   local slug="${1:?usage: down <slug> --task-root <dir>}"; shift
   local task_root="" keep_db=0 hexer_port=""
   while [ $# -gt 0 ]; do
      case "$1" in
         --task-root)  task_root="$2"; shift 2 ;;
         --keep-db)    keep_db=1; shift ;;
         --hexer-port) hexer_port="$2"; shift 2 ;;
         *) die "unknown down option: $1" ;;
      esac
   done
   local name; name="$(container_name "$slug")"

   if [ -n "$task_root" ] && [ -f "$task_root/.hexer/hexer.pid" ]; then
      local pid; pid="$(cat "$task_root/.hexer/hexer.pid")"
      if kill -0 "$pid" 2>/dev/null; then
         kill "$pid" 2>/dev/null || true
         ok "hexer (pid $pid) stopped"
      fi
      rm -f "$task_root/.hexer/hexer.pid"
   fi
   if [ -n "$task_root" ] && [ -f "$task_root/.hexer/tls.pid" ]; then
      kill "$(cat "$task_root/.hexer/tls.pid")" 2>/dev/null || true
      rm -f "$task_root/.hexer/tls.pid"
   fi
   # The pidfile can go stale/wrong; the ports are the source of truth.
   if [ -n "$hexer_port" ]; then
      kill_port_listeners "$hexer_port"
      kill_port_listeners "$((hexer_port + 1000))"
   fi
   if container_exists "$name"; then
      if [ "$keep_db" = "1" ]; then
         docker stop "$name" >/dev/null 2>&1 || true
         ok "db container $name stopped (kept)"
      else
         docker rm -f "$name" >/dev/null
         ok "db container $name removed"
      fi
   fi
}

cmd_status() {
   local slug="${1:?usage: status <slug> --task-root <dir>}"; shift
   local task_root=""
   while [ $# -gt 0 ]; do
      case "$1" in
         --task-root) task_root="$2"; shift 2 ;;
         *) die "unknown status option: $1" ;;
      esac
   done
   local name; name="$(container_name "$slug")"
   local hexer="stopped"
   if [ -n "$task_root" ] && [ -f "$task_root/.hexer/hexer.pid" ]; then
      kill -0 "$(cat "$task_root/.hexer/hexer.pid")" 2>/dev/null && hexer="running"
   fi
   local db="absent"
   if container_running "$name"; then db="running"
   elif container_exists "$name"; then db="stopped"
   fi
   echo "hexer=$hexer db=$db"
}

# ── setup ─────────────────────────────────────────────────────────────────────
# One-shot machine bootstrap for new developers: clone the repos when
# missing, pull the Oracle image, install hexer deps, generate the cert and
# build the first golden image. Idempotent — done steps are skipped. Emits
# the same step/pct progress lines as `up`.
cmd_setup() {
   local hexer_dir="" tds_dir="" branch="master" suffix="acrid.dev"
   local hexer_repo="git@github.com:acreidentity/tds-hexer.git"
   local tds_repo="git@github.com:acreidentity/tds-suite.git"
   while [ $# -gt 0 ]; do
      case "$1" in
         --hexer-dir)  hexer_dir="$2"; shift 2 ;;
         --tds-dir)    tds_dir="$2"; shift 2 ;;
         --hexer-repo) hexer_repo="$2"; shift 2 ;;
         --tds-repo)   tds_repo="$2"; shift 2 ;;
         --branch)     branch="$2"; shift 2 ;;
         --suffix)     suffix="$2"; shift 2 ;;
         *) die "unknown setup option: $1" ;;
      esac
   done
   [ -n "$hexer_dir" ] && [ -n "$tds_dir" ] || die "setup needs --hexer-dir and --tds-dir"
   HEXER_DEFAULT="$hexer_dir"
   TDS="$tds_dir"

   PROVISION_LOG="$HOME/.config/acorn/logs/hexer-setup.log"
   mkdir -p "$(dirname "$PROVISION_LOG")"
   : > "$PROVISION_LOG"
   exec 2> >(tee -a "$PROVISION_LOG" >&2)

   step validate 3 "checking base tools"
   command -v git >/dev/null    || die "git not installed"
   command -v node >/dev/null   || die "node not installed (need >=22): brew install node"
   command -v docker >/dev/null || die "docker not installed: install Docker Desktop"
   docker info >/dev/null 2>&1  || die "Docker daemon is not running — start Docker Desktop"
   ensure_pnpm

   step clone-hexer 8 "hexer checkout"
   if [ ! -d "$hexer_dir/.git" ]; then
      mkdir -p "$(dirname "$hexer_dir")"
      GIT_TERMINAL_PROMPT=0 git clone "$hexer_repo" "$hexer_dir" \
         || die "cannot clone $hexer_repo — is your GitHub SSH key set up?"
      ok "cloned $hexer_repo"
   else
      ok "hexer checkout already present"
   fi

   step clone-tds 18 "tds-suite checkout (large repo)"
   if [ ! -d "$tds_dir/.git" ]; then
      mkdir -p "$(dirname "$tds_dir")"
      GIT_TERMINAL_PROMPT=0 git clone "$tds_repo" "$tds_dir" \
         || die "cannot clone $tds_repo — is your GitHub SSH key set up?"
      ok "cloned $tds_repo"
   else
      ok "tds-suite checkout already present"
   fi

   step pull-image 30 "oracle image (~2.5 GB, skipped when present)"
   if ! docker image inspect "$ORACLE_IMAGE" >/dev/null 2>&1; then
      docker pull "$ORACLE_IMAGE" >&2 || die "docker pull failed for $ORACLE_IMAGE"
   fi

   step hexer-deps 55 "pnpm install in hexer"
   if [ ! -d "$hexer_dir/node_modules" ]; then
      ( cd "$hexer_dir" && pnpm install ) >&2 || die "pnpm install failed in $hexer_dir"
   fi

   step cert 65 "dev certificate for *.$suffix"
   ensure_dev_cert "$suffix"

   step build-golden 70 "golden DB image for $branch (30-60 min first time)"
   local tag; tag="$(golden_tag "$branch")"
   if ! docker image inspect "$GOLDEN_REPO:$tag" >/dev/null 2>&1; then
      cmd_build_golden "$branch"
   fi

   if cert_is_trusted; then
      step done 100 "setup complete — create a task with 'enable hexer'"
   else
      step done 100 "one manual step left: $(cert_trust_command)"
   fi
}

# ── dispatch ──────────────────────────────────────────────────────────────────
cmd="${1:-}"
[ -n "$cmd" ] || { sed -n '3,22p' "$0" >&2; exit 1; }
shift
case "$cmd" in
   doctor)       cmd_doctor "$@" ;;
   setup)        cmd_setup "$@" ;;
   build-golden) cmd_build_golden "$@" ;;
   up)           cmd_up "$@" ;;
   down)         cmd_down "$@" ;;
   status)       cmd_status "$@" ;;
   -h|--help)    sed -n '3,22p' "$0"; exit 0 ;;
   *)            die "unknown command: $cmd (doctor | build-golden | up | down | status)" ;;
esac
