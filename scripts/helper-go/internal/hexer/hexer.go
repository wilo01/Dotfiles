// Package hexer provisions a task's dedicated TDS environment: an Oracle
// container cloned from a version-matched golden image plus a Hexer process
// serving the task's tds-suite worktree.
//
// The provisioning itself lives in hexer-task.sh, which is embedded in the
// binary and materialized next to the config. This package only drives it and
// renders its progress.
package hexer

import (
	"bufio"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

//go:embed hexer-task.sh
var script []byte

// Module is one hexer-served app within tds-suite. The script takes these as
// name:route:static:rt quadruples.
type Module struct {
	Name   string `mapstructure:"name" yaml:"name"`
	Route  string `mapstructure:"route" yaml:"route"`
	Static string `mapstructure:"static" yaml:"static"`
	RT     string `mapstructure:"rt" yaml:"rt"`
}

func (m Module) arg() string {
	return strings.Join([]string{m.Name, m.Route, m.Static, m.RT}, ":")
}

// DefaultModules mirror the apps claudiusz ships with, used when the config
// names none.
func DefaultModules() []Module {
	return []Module{
		{Name: "safe", Route: "safe", Static: "source/ui", RT: "source/server/rt"},
		{Name: "kiosk", Route: "kiosk", Static: "source/ui-kiosk", RT: "source/server/rtkiosk"},
	}
}

// Config is the machine-level hexer setup, from agent.hexer in config.yaml.
type Config struct {
	Enabled        bool     `mapstructure:"enabled" yaml:"enabled"`
	HexerDir       string   `mapstructure:"hexer_dir" yaml:"hexer_dir"`
	TDSDir         string   `mapstructure:"tds_dir" yaml:"tds_dir"`
	TDSRepo        string   `mapstructure:"tds_repo" yaml:"tds_repo"`
	HexerPortMin   int      `mapstructure:"hexer_port_min" yaml:"hexer_port_min"`
	HexerPortMax   int      `mapstructure:"hexer_port_max" yaml:"hexer_port_max"`
	DBPortMin      int      `mapstructure:"db_port_min" yaml:"db_port_min"`
	DBPortMax      int      `mapstructure:"db_port_max" yaml:"db_port_max"`
	HostnameSuffix string   `mapstructure:"hostname_suffix" yaml:"hostname_suffix"`
	Modules        []Module `mapstructure:"modules" yaml:"modules"`
}

// EffectiveModules falls back to the built-in app list when none is configured.
func (c Config) EffectiveModules() []Module {
	if len(c.Modules) > 0 {
		return c.Modules
	}
	return DefaultModules()
}

// ModuleNames lists the selectable app names for the picker.
func (c Config) ModuleNames() []string {
	mods := c.EffectiveModules()
	names := make([]string, 0, len(mods))
	for _, m := range mods {
		names = append(names, m.Name)
	}
	return names
}

// ModulesByName resolves selected names back to full module definitions,
// erroring on a name the config does not define.
func (c Config) ModulesByName(names []string) ([]Module, error) {
	byName := map[string]Module{}
	for _, m := range c.EffectiveModules() {
		byName[m.Name] = m
	}
	var out []Module
	for _, n := range names {
		m, ok := byName[n]
		if !ok {
			return nil, fmt.Errorf("unknown hexer app %q (configured: %s)", n, strings.Join(c.ModuleNames(), ", "))
		}
		out = append(out, m)
	}
	return out, nil
}

// AllocatePort returns the lowest free port in [min,max] that is neither
// claimed by another task nor already listening on this machine.
func AllocatePort(min, max int, used map[int]bool) (int, error) {
	for port := min; port <= max; port++ {
		if used[port] {
			continue
		}
		return port, nil
	}
	return 0, fmt.Errorf("no free port in range %d-%d", min, max)
}

// Hostname is the Host-header name the task's hexer answers on.
func Hostname(key, suffix string) string {
	return strings.ToLower(key) + "." + suffix
}

// EnsureScript materializes the embedded script at dir/hexer-task.sh, rewriting
// it whenever the on-disk copy differs from the embedded one so a hlp upgrade
// cannot leave a stale script behind.
func EnsureScript(dir string) (string, error) {
	path := filepath.Join(dir, "hexer-task.sh")
	want := sha256.Sum256(script)

	if existing, err := os.ReadFile(path); err == nil {
		if got := sha256.Sum256(existing); hex.EncodeToString(got[:]) == hex.EncodeToString(want[:]) {
			return path, nil
		}
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create %s: %w", dir, err)
	}
	if err := os.WriteFile(path, script, 0o755); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", path, err)
	}
	return path, nil
}

// Runner drives the script for one machine configuration.
type Runner struct {
	ScriptPath string
	Config     Config
	// Progress receives step name, percent and message as the script reports
	// them. Nil discards progress.
	Progress func(step string, pct int, msg string)
}

// New materializes the script and returns a runner for it.
func New(configDir string, cfg Config, progress func(string, int, string)) (*Runner, error) {
	path, err := EnsureScript(configDir)
	if err != nil {
		return nil, err
	}
	return &Runner{ScriptPath: path, Config: cfg, Progress: progress}, nil
}

// UpOptions describes the environment to bring up for one task.
type UpOptions struct {
	Slug      string
	Branch    string
	Worktree  string
	TaskRoot  string
	DBPort    int
	HexerPort int
	Host      string
	Modules   []Module
}

// Up provisions the DB container and hexer process for a task.
func (r *Runner) Up(opts UpOptions) error {
	args := []string{"up", opts.Slug,
		"--branch", opts.Branch,
		"--worktree", opts.Worktree,
		"--task-root", opts.TaskRoot,
		"--db-port", strconv.Itoa(opts.DBPort),
		"--hexer-port", strconv.Itoa(opts.HexerPort),
		"--build-golden-if-missing",
	}
	if opts.Host != "" {
		args = append(args, "--host", opts.Host)
	}
	for _, m := range opts.Modules {
		args = append(args, "--module", m.arg())
	}
	return r.run(args...)
}

// Down tears the environment down. Called before worktrees are removed, because
// the container is bound to the worktree path.
func (r *Runner) Down(slug, taskRoot string, hexerPort int, keepDB bool) error {
	args := []string{"down", slug, "--task-root", taskRoot}
	if hexerPort != 0 {
		args = append(args, "--hexer-port", strconv.Itoa(hexerPort))
	}
	if keepDB {
		args = append(args, "--keep-db")
	}
	return r.run(args...)
}

// Status reports the environment's state for one task.
func (r *Runner) Status(slug, taskRoot string) (string, error) {
	cmd := r.command("status", slug, "--task-root", taskRoot)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// Doctor runs the blocking subset of environment pre-checks. It is run before
// Up so a missing docker or JDK fails immediately rather than halfway through
// provisioning.
func (r *Runner) Doctor(branch string) error {
	cmd := r.command("doctor", "--gate", branch)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("hexer environment is not ready:\n%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// command builds the script invocation. Machine paths reach the script through
// the environment, never from its own location, because it is materialized away
// from the repos it operates on.
func (r *Runner) command(args ...string) *exec.Cmd {
	cmd := exec.Command(r.ScriptPath, args...)
	cmd.Env = append(os.Environ(),
		"HEXER_TASK_HEXER_DIR="+expand(r.Config.HexerDir),
		"HEXER_TASK_TDS_DIR="+expand(r.Config.TDSDir),
	)
	return cmd
}

// run streams the script's stdout, turning its step= lines into progress
// callbacks, and lets its human-readable stderr through untouched.
func (r *Runner) run(args ...string) error {
	cmd := r.command(args...)
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to capture hexer-task output: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start hexer-task.sh: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		step, pct, msg, ok := parseProgress(scanner.Text())
		if !ok {
			fmt.Println(scanner.Text())
			continue
		}
		if r.Progress != nil {
			r.Progress(step, pct, msg)
		}
	}
	io.Copy(io.Discard, stdout)

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("hexer-task %s failed: %w", args[0], err)
	}
	return nil
}

// parseProgress reads a "step=<name> pct=<n> msg=<text>" line. msg runs to the
// end of the line, so it is taken last rather than split on spaces.
func parseProgress(line string) (step string, pct int, msg string, ok bool) {
	rest, found := strings.CutPrefix(line, "step=")
	if !found {
		return "", 0, "", false
	}
	step, rest, found = strings.Cut(rest, " pct=")
	if !found {
		return "", 0, "", false
	}
	pctText := rest
	if before, after, hasMsg := strings.Cut(rest, " msg="); hasMsg {
		pctText, msg = before, after
	}
	pct, err := strconv.Atoi(strings.TrimSpace(pctText))
	if err != nil {
		return "", 0, "", false
	}
	return step, pct, msg, true
}

func expand(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, strings.TrimPrefix(path, "~"))
	}
	return path
}
