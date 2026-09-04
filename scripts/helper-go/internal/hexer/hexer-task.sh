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
#   down <slug> --task-root <dir> [--db-port <n>] [--keep-db]
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
# Basic-auth credentials for hexer's /health dashboard. /health-check itself is
# unauthenticated, so provisioning never needs these.
HEALTH_USER="${HEXER_TASK_HEALTH_USER:-admin}"
HEALTH_PASS="${HEXER_TASK_HEALTH_PASS:-admin123}"
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

# A task container is cloned from an image that was committed while its database
# was already up, so its log carries that build's "DATABASE IS READY TO USE"
# line from the very first poll. Matching it is what let liquibase start against
# a PDB that had not reopened yet (ORA-12514). The healthcheck is authoritative
# where there is one; the log is consulted only when there is not, and then only
# for output this container produced since it started.
wait_db_ready() {
   local name="$1" tries="${2:-240}"
   local i health started
   local -a since=()
   started="$(docker inspect -f '{{.State.StartedAt}}' "$name" 2>/dev/null || true)"
   if [ -n "$started" ]; then
      since=(--since "$started")
   fi
   for i in $(seq 1 "$tries"); do
      health="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{end}}' "$name" 2>/dev/null || true)"
      if [ -n "$health" ]; then
         [ "$health" = "healthy" ] && return 0
      # grep without -q: it must consume the whole stream — early exit sends
      # docker logs a SIGPIPE, which pipefail turns into a false negative.
      elif docker logs "${since[@]}" "$name" 2>&1 | grep "DATABASE IS READY TO USE" >/dev/null; then
         return 0
      fi
      container_running "$name" || die "container $name stopped — check 'docker logs $name'"
      sleep 5
   done
   die "$name did not become ready — check 'docker logs $name' (Rosetta enabled?)"
}

CERT_DIR="${HEXER_TASK_CERT_DIR:-$HOME/.config/hlp/certs}"

# The .dev TLD is HSTS-preloaded — browsers force HTTPS, so plain HTTP can
# never work and the certificate has to be one the browser accepts.
#
# tds-hexer ships its own dev CA (scripts/dev/generate-dev-certs.sh, exposed as
# `pnpm dev:certs`) covering *.acrid.dev. Reusing it means one import into the
# browser covers every task environment; generating a cert per suffix here would
# mean a new untrusted authority each time.
hexer_repo_cert() {
   local dir="${HEXER_DEFAULT:-}"
   [ -n "$dir" ] || return 1
   HEXER_CERT_DIR="$dir/src/assets/certs/dev"
   [ -f "$HEXER_CERT_DIR/acreidentity-dev.crt" ] && [ -f "$HEXER_CERT_DIR/acreidentity-dev.key" ]
}

ensure_dev_cert() {
   local suffix="$1"

   if hexer_repo_cert; then
      CERT_FILE="$HEXER_CERT_DIR/acreidentity-dev.crt"
      KEY_FILE="$HEXER_CERT_DIR/acreidentity-dev.key"
      ok "using the tds-hexer dev certificate ($CERT_FILE)"
      return 0
   fi

   # Not generated yet: let the repo's own tooling make it, so the cert the
   # browser trusts and the cert we serve stay the same artifact.
   if [ -n "${HEXER_DEFAULT:-}" ] && [ -f "$HEXER_DEFAULT/scripts/dev/generate-dev-certs.sh" ]; then
      warn "tds-hexer dev certificate missing — generating it with pnpm dev:certs"
      ( cd "$HEXER_DEFAULT" && pnpm dev:certs ) >&2 2>&1 || true
      if hexer_repo_cert; then
         CERT_FILE="$HEXER_CERT_DIR/acreidentity-dev.crt"
         KEY_FILE="$HEXER_CERT_DIR/acreidentity-dev.key"
         ok "generated the tds-hexer dev certificate ($CERT_FILE)"
         return 0
      fi
   fi

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
   warn "no tds-hexer dev certificate found; fell back to a self-signed cert for *.$suffix"
   ok "self-signed wildcard cert created ($CERT_DIR)"
}

# HSTS-preloaded TLDs (.dev) hard-block untrusted certs — no "Proceed
# anyway" in Chrome. The cert must be a trusted anchor in the System
# keychain. Deliberately NOT automated: changing the system trust store is
# the developer's own call, so we detect and hand them the exact command.
cert_trust_command() {
   if [ "$(uname)" = "Darwin" ]; then
      echo "sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain '$CERT_FILE'"
   else
      echo "certutil -d sql:\$HOME/.pki/nssdb -A -t 'C,,' -n 'Acre Identity Dev' -i '$CERT_FILE'"
   fi
}

# Chrome and Firefox read the NSS database rather than the system store, so on
# Linux that is where the answer is. An unknown answer counts as trusted: this
# only drives a hint, and a false alarm on every run would be worse than silence.
cert_is_trusted() {
   if [ "$(uname)" = "Darwin" ]; then
      security verify-cert -c "$CERT_FILE" -p ssl >/dev/null 2>&1
      return
   fi
   command -v certutil >/dev/null 2>&1 || return 0
   [ -d "$HOME/.pki/nssdb" ] || return 0
   certutil -d "sql:$HOME/.pki/nssdb" -L 2>/dev/null | grep -q "Acre Identity Dev"
}

CERT_TRUST_MISSING=""

ensure_cert_trusted() {
   cert_is_trusted && return 0
   CERT_TRUST_MISSING="1"
   warn "dev certificate is not trusted yet — the environment will still start, but the browser will refuse it"
   return 0
}

# cert_trust_reminder repeats the one-time import next to the URL it unblocks.
cert_trust_reminder() {
   [ -n "$CERT_TRUST_MISSING" ] || return 0
   warn "trust the dev certificate once, then reopen the URL above:"
   warn "  $(cert_trust_command)"
   warn "  Chrome: chrome://certificate-manager/localcerts/usercerts → Import → $CERT_FILE"
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

# True when every address the resolver returns for $host is loopback. Checking
# resolution rather than /etc/hosts is what lets a wildcard DNS entry, or
# systemd-resolved synthesising *.localhost, satisfy this without root.
resolves_to_loopback() {
   local host="$1" addrs=""
   if command -v getent >/dev/null 2>&1; then
      addrs="$(getent ahosts "$host" 2>/dev/null | awk '{print $1}' | sort -u)"
   elif command -v dscacheutil >/dev/null 2>&1; then
      addrs="$(dscacheutil -q host -a name "$host" 2>/dev/null | awk '/^ipv?6?_address:/ {print $2}' | sort -u)"
   fi
   [ -n "$addrs" ] || return 1

   local addr
   while IFS= read -r addr; do
      [ -n "$addr" ] || continue
      case "$addr" in
         127.*|::1) ;;
         # A public record pointing somewhere real must not be mistaken for a
         # local mapping.
         *) return 1 ;;
      esac
   done <<RESOLVED
$addrs
RESOLVED
   return 0
}

# Set when the task hostname does not resolve, so cmd_up can repeat the fix at
# the end rather than burying it behind the provisioning log.
HOST_MAPPING_MISSING=""

# Task hostnames (<slug>.<suffix>) must resolve to loopback for a browser to
# reach the environment. Nothing in provisioning needs it — the DB and hexer
# bind to ports, not names — so a missing mapping is reported, never fatal.
ensure_host_mapping() {
   local host="$1"
   resolves_to_loopback "$host" && return 0

   if [ "$(uname)" = "Darwin" ]; then
      warn "$host does not resolve — requesting admin approval to add it to /etc/hosts"
      if osascript -e "do shell script \"printf '127.0.0.1 $host\\n' >> /etc/hosts\" with administrator privileges" >/dev/null 2>&1; then
         ok "$host → 127.0.0.1 added to /etc/hosts"
         return 0
      fi
   fi

   HOST_MAPPING_MISSING="$host"
   warn "$host does not resolve yet — continuing; the environment will still come up"
   warn "run this when convenient:  sudo sh -c 'echo 127.0.0.1 $host >> /etc/hosts'"
   return 0
}

# host_mapping_reminder repeats the pending /etc/hosts line next to the URL it
# unblocks, so it is the last thing on screen instead of the first.
host_mapping_reminder() {
   [ -n "$HOST_MAPPING_MISSING" ] || return 0
   warn "$HOST_MAPPING_MISSING is not in /etc/hosts yet — the URL above will not resolve until it is:"
   warn "  sudo sh -c 'echo 127.0.0.1 $HOST_MAPPING_MISSING >> /etc/hosts'"
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
# Personal seed data (custom_default_data.sql and friends) is deliberately
# gitignored, so it lives only in the shared checkout and never reaches a task
# worktree -- where safe.changelog.xml's `includeAll after-install/` would
# otherwise pick it up. Linking rather than copying keeps the shared checkout
# the single place to edit it. Only ignored files are linked: a tracked
# after-install file already exists in the worktree at that branch's version.
link_personal_after_install() {
   local worktree="$1"
   local rel="source/server/database/sql/safe/after-install"
   local src="$TDS/$rel" dst="$worktree/$rel"
   [ -d "$src" ] && [ -d "$dst" ] || return 0
   local f base
   while IFS= read -r f; do
      [ -n "$f" ] || continue
      base="$(basename "$f")"
      ln -sfn "$src/$base" "$dst/$base"
      ok "seeding $base from the shared checkout"
   done < <( cd "$TDS" && git ls-files --others --ignored --exclude-standard -- "$rel" 2>/dev/null )
}

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
     local rc=0
     ( cd source/server/database && rm -f install-update.log && JAVA_HOME="$LB_JAVA_HOME" ./liquibase update ) || rc=$?
     git restore source/server/database/liquibase.properties source/server/database/sql/safe/safe.changelog.xml 2>/dev/null || true
     [ "$rc" -eq 0 ] || die "liquibase update failed (exit $rc) — see $worktree/source/server/database/install-update.log"
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

# ── saved database connections ────────────────────────────────────────────────
# A task's DB is only reachable at localhost:<db-port>, so every task needs its
# own entry in the tools that talk to it: SQLcl's common connection store
# (~/.dbtools, also read by the oracle-sqlcl MCP and SQL Developer for VS Code)
# and SQL Developer desktop's own store.

connection_name() { echo "oracle-docker-$1"; }

sqlcl_bin() {
   local candidate
   for candidate in "${HEXER_TASK_SQLCL:-}" "${SQLCL_PATH:-}" \
                    "$(command -v sql 2>/dev/null || true)" \
                    "$HOME/.local/sqlcl/bin/sql" /opt/sqlcl/bin/sql; do
      [ -n "$candidate" ] && [ -x "$candidate" ] && { echo "$candidate"; return 0; }
   done
   return 1
}

# `connect -save` only persists on a successful connect, so this doubles as a
# check that the schema is actually reachable with the credentials hexer uses.
save_sqlcl_connection() {
   local db_port="$1" name bin
   name="$(connection_name "$db_port")"
   bin="$(sqlcl_bin)" || { warn "sqlcl not found — skipping the $name connection (set HEXER_TASK_SQLCL)"; return 1; }
   printf 'set feedback off\nconnect -save %s -savepwd -replace %s/%s@localhost:%s/%s\nexit\n' \
      "$name" "$DB_USER" "$DB_PASS" "$db_port" "$DB_SERVICE" \
      | "$bin" -nohistory /nolog >/dev/null 2>&1
}

remove_sqlcl_connection() {
   local name bin
   name="$(connection_name "$1")"
   bin="$(sqlcl_bin)" || return 0
   printf 'set feedback off\nconnmgr delete -conn %s\nexit\n' "$name" \
      | "$bin" -nohistory /nolog >/dev/null 2>&1 || true
}

sqldev_stores() {
   ls -d "$HOME"/.sqldeveloper/system*/o.jdeveloper.db.connection/connections.json 2>/dev/null || true
}

sqldev_running() {
   pgrep -f 'sqldeveloper.*ide-launcher|oracle.ide.boot.Launcher' >/dev/null 2>&1
}

# SQL Developer desktop encrypts passwords with a per-connection key derived
# from its own install, which cannot be reproduced here — the entry is written
# without one, so the IDE prompts on first connect. It also rewrites this file
# wholesale on exit, so a running instance would silently discard the edit.
sqldev_write_connection() {
   local db_port="$1" action="$2" name store rc=0
   name="$(connection_name "$db_port")"
   [ -n "$(sqldev_stores)" ] || return 0
   if sqldev_running; then
      warn "SQL Developer is running — close it and re-run to get the $name connection there"
      return 0
   fi
   for store in $(sqldev_stores); do
      run_node_script "$store" "$name" "$db_port" "$DB_USER" "$DB_SERVICE" "$action" <<'NODE' || rc=$?
const fs = require("fs");
const [store, name, port, user, service, action] = process.argv.slice(2);
const doc = JSON.parse(fs.readFileSync(store, "utf8"));
const kept = (doc.connections || []).filter((c) => c.name !== name);
if (action === "add") {
  kept.push({
    name,
    type: "jdbc",
    info: {
      role: "", SavePassword: "false", NoPasswordConnection: "TRUE",
      OracleConnectionType: "BASIC", RaptorConnectionType: "Oracle",
      subtype: "oraJDBC", driver: "oracle.jdbc.OracleDriver", oraDriverType: "thin",
      hostname: "localhost", port, serviceName: service, user,
      customUrl: `jdbc:oracle:thin:@//localhost:${port}/${service}`,
      OS_AUTHENTICATION: "false", KERBEROS_AUTHENTICATION: "false",
      IS_PROXY: "false", PROXY_TYPE: "USER NAME", PROXY_USER_NAME: "",
      ConnName: name,
    },
  });
}
doc.connections = kept;
fs.copyFileSync(store, `${store}.hexer-task.bak`);
fs.writeFileSync(store, JSON.stringify(doc, null, 2));
NODE
   done
   sqldev_write_folder "$db_port" "$action" || warn "could not file $name under $SQLDEV_FOLDER in SQL Developer"
   return $rc
}

# Folder membership lives outside connections.json, in a flat name→folder map
# in product-preferences.xml. Without an entry there the connection still works
# but lands at the root of the navigator instead of alongside its siblings.
SQLDEV_FOLDER="${HEXER_TASK_SQLDEV_FOLDER:-Dockers}"

sqldev_prefs() {
   ls -d "$HOME"/.sqldeveloper/system*/o.sqldeveloper/product-preferences.xml 2>/dev/null || true
}

sqldev_write_folder() {
   local db_port="$1" action="$2" name prefs rc=0
   name="$(connection_name "$db_port")"
   for prefs in $(sqldev_prefs); do
      run_node_script "$prefs" "$name" "$SQLDEV_FOLDER" "$action" <<'NODE' || rc=$?
const fs = require("fs");
const [prefs, name, folder, action] = process.argv.slice(2);

const formEncode = (s) =>
  encodeURIComponent(s).replace(/%20/g, "+").replace(/[!'()*]/g, (c) =>
    "%" + c.charCodeAt(0).toString(16).toUpperCase());

const lines = fs.readFileSync(prefs, "utf8").split("\n");
const openTag = (i, n) => lines[i].trim().startsWith(`<hash n="${n}">`);

const descend = (from, to, names) => {
  let start = from;
  for (const n of names) {
    let found = -1;
    for (let i = start; i < to; i++) if (openTag(i, n)) { found = i; break; }
    if (found === -1) return null;
    start = found + 1;
  }
  let depth = 0;
  for (let i = start; i < to; i++) {
    const t = lines[i].trim();
    if (t.startsWith("<hash")) depth++;
    else if (t.startsWith("</hash>")) { if (depth === 0) return [start, i]; depth--; }
  }
  return null;
};

const span = descend(0, lines.length, ["DatabaseFoldersCache", "Connections", "sqldev.nav"]);
if (!span) process.exit(3);
const [first, end] = span;

const key = `IdeConnections%23${formEncode(name)}`;
const body = lines.slice(first, end).filter((l) => !l.includes(`n="${key}"`));

if (action === "add") {
  const indent = (lines[first] || "").match(/^\s*/)[0] || "            ";
  const entry = `${indent}<value n="${key}" v="${formEncode(folder)}"/>`;
  const at = body.findIndex((l) => {
    const m = l.match(/n="([^"]+)"/);
    return m && m[1].localeCompare(key) > 0;
  });
  body.splice(at === -1 ? body.length : at, 0, entry);
}

fs.copyFileSync(prefs, `${prefs}.hexer-task.bak`);
fs.writeFileSync(prefs, [...lines.slice(0, first), ...body, ...lines.slice(end)].join("\n"));
NODE
   done
   return $rc
}

register_db_connections() {
   local db_port="$1" name
   name="$(connection_name "$db_port")"
   if save_sqlcl_connection "$db_port"; then
      ok "saved connection $name (localhost:$db_port/$DB_SERVICE as $DB_USER)"
   else
      warn "could not save the $name connection to the SQLcl store"
   fi
   sqldev_write_connection "$db_port" add || warn "could not add $name to SQL Developer desktop"
}

unregister_db_connections() {
   remove_sqlcl_connection "$1"
   sqldev_write_connection "$1" remove || true
   ok "removed connection $(connection_name "$1")"
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
   link_personal_after_install "$worktree"
   run_liquibase "$worktree" "$db_port"
   recompile_schema "$name"
   step db-connection 70 "saving the $(connection_name "$db_port") database connection"
   register_db_connections "$db_port"

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
const { existsSync } = require("fs");
const [slug, host, worktree, user, pass, port, service, ...moduleSpecs] = process.argv.slice(2);

// A Sencha Cmd app serves a microloader that needs bootstrap.js/bootstrap.json,
// which are build output and gitignored -- so a fresh worktree 404s them and the
// page dies on "Ext is not defined". app.json is what marks such an app; hexer's
// startup-command-service runs this in the module's static dir and keeps it there.
const startupCommandFor = (staticDir) =>
  existsSync(`${staticDir}/app.json`) ? "sencha app watch" : undefined;

const modules = moduleSpecs.map((spec) => {
  const [name, route, staticRel, rtRel] = spec.split(":");
  const staticDir = `${worktree}/${staticRel}`;
  return {
    appName: name,
    type: "web",
    version: "local",
    route: `/${route.replace(/^\//, "")}`,
    static: staticDir,
    startupCommand: startupCommandFor(staticDir),
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
   write_health_env "$hexer_state"
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
     HEALTH_AUTH_USERNAME="$HEALTH_USER" HEALTH_AUTH_PASSWORD="$HEALTH_PASS" \
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

   warm_build_watchers "$worktree" "$host" "$internal_port" "${modules[@]}"

   step done 100 "https://$host:$hexer_port"
   ok "task environment up: https://$host:$hexer_port (db localhost:$db_port)"
   host_mapping_reminder
   cert_trust_reminder
}

# Hexer starts a module's build watcher lazily, on the first static request to it
# (startup-command-service's ensureStartupCommand), and Sencha's first build takes
# about a minute. Touching the route here means that build is already under way --
# usually finished -- by the time the page is opened, rather than the first visit
# 404ing bootstrap.js and dying on "Ext is not defined".
# Hexer runs with the task state dir as its cwd, so dotenv reads a file placed
# there. The launch below already exports these, which covers every hexer this
# script starts; the file additionally covers one restarted by hand from that
# directory. Only the two keys are rewritten, so anything else in the file (and
# INSTANCES, which is never written here) survives.
write_health_env() {
   local hexer_state="$1"
   local envfile="$hexer_state/.env"
   local tmp
   tmp="$(mktemp)"
   if [ -f "$envfile" ]; then
      grep -vE '^(HEALTH_AUTH_USERNAME|HEALTH_AUTH_PASSWORD)=' "$envfile" > "$tmp" || true
   fi
   printf 'HEALTH_AUTH_USERNAME=%s\nHEALTH_AUTH_PASSWORD=%s\n' "$HEALTH_USER" "$HEALTH_PASS" >> "$tmp"
   install -m 600 "$tmp" "$envfile"
   rm -f "$tmp"
}

warm_build_watchers() {
   local worktree="$1" host="$2" internal_port="$3"
   shift 3
   local spec name route static_rel
   for spec in "$@"; do
      IFS=: read -r name route static_rel _ <<< "$spec"
      [ -f "$worktree/$static_rel/app.json" ] || continue
      curl -fsS --max-time 10 -H "Host: $host" \
         "http://localhost:$internal_port/$route/static/" >/dev/null 2>&1 || true
      ok "$name build watcher started (first build takes ~1 min)"
   done
}

# ── down / status ─────────────────────────────────────────────────────────────
cmd_down() {
   local slug="${1:?usage: down <slug> --task-root <dir>}"; shift
   local task_root="" keep_db=0 hexer_port="" db_port=""
   while [ $# -gt 0 ]; do
      case "$1" in
         --task-root)  task_root="$2"; shift 2 ;;
         --keep-db)    keep_db=1; shift ;;
         --hexer-port) hexer_port="$2"; shift 2 ;;
         --db-port)    db_port="$2"; shift 2 ;;
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
         if [ -n "$db_port" ]; then unregister_db_connections "$db_port"; fi
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
