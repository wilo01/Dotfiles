package agent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dariuszw/hlp/internal/config"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

// certNickname is the name tds-hexer's generator gives the dev CA, and the one
// its own removal instructions use.
const certNickname = "Acre Identity Dev"

// certRelPath is where tds-hexer's `pnpm dev:certs` writes the CA.
var certRelPath = filepath.Join("src", "assets", "certs", "dev", "acreidentity-dev.crt")

var (
	certStatus bool
	certRemove bool
)

var certCmd = &cobra.Command{
	Use:   "cert",
	Short: "Trust the tds-hexer dev certificate so task environments open without a warning",
	Long: `Imports tds-hexer's dev CA into every certificate store on this machine.

The CA covers *.acrid.dev, so one import makes every task environment open
cleanly. Chrome and Firefox each keep their own NSS store, and neither is
covered by the OS trust commands, so both are handled here.

Re-running is safe: existing copies are replaced rather than stacked, which
also cleans up duplicates from earlier manual imports.

  hlp agent cert            import into every store
  hlp agent cert --status   report where it is trusted
  hlp agent cert --remove   remove it from every store`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		return runCert(cfg)
	},
}

// certStore is one NSS database certutil can operate on.
type certStore struct {
	Label string
	Dir   string
}

func runCert(cfg *config.Config) error {
	if runtime.GOOS == "darwin" {
		return certMacOS(cfg)
	}
	if _, err := exec.LookPath("certutil"); err != nil {
		return fmt.Errorf("certutil not found — install it first (Fedora: sudo dnf install nss-tools)")
	}

	stores := discoverCertStores()
	if len(stores) == 0 {
		return fmt.Errorf("no certificate store found (no ~/.pki/nssdb and no Firefox profile)")
	}

	if certStatus {
		return reportCertStores(stores)
	}
	if certRemove {
		for _, store := range stores {
			removed := removeCertFrom(store)
			fmt.Println(ui.SuccessMsg(fmt.Sprintf("%s: removed %d copy/copies", store.Label, removed)))
		}
		return nil
	}

	certPath, err := hexerCertPath(cfg)
	if err != nil {
		return err
	}
	fmt.Println(ui.KeyValue("Certificate", certPath))

	for _, store := range stores {
		if err := installCertInto(store, certPath); err != nil {
			fmt.Println(ui.Warning(store.Label + ": " + err.Error()))
			continue
		}
		fmt.Println(ui.SuccessMsg(store.Label + ": trusted"))
	}

	fmt.Println(ui.Info("restart the browser for it to pick the change up"))
	return nil
}

// discoverCertStores finds the NSS databases in play: Chrome's shared one and
// every Firefox profile, since Firefox ignores the shared store entirely.
func discoverCertStores() []certStore {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	var stores []certStore
	if shared := filepath.Join(home, ".pki", "nssdb"); dirExists(shared) {
		stores = append(stores, certStore{Label: "Chrome/Chromium", Dir: shared})
	}

	profiles, _ := filepath.Glob(filepath.Join(home, ".mozilla", "firefox", "*", "cert9.db"))
	for _, profile := range profiles {
		dir := filepath.Dir(profile)
		stores = append(stores, certStore{Label: "Firefox (" + filepath.Base(dir) + ")", Dir: dir})
	}
	return stores
}

// hexerCertPath locates the CA, generating it with the repo's own tooling when
// it has not been made yet.
func hexerCertPath(cfg *config.Config) (string, error) {
	hexerDir := expandPath(cfg.Agent.Hexer.HexerDir)
	if hexerDir == "" {
		return "", fmt.Errorf("agent.hexer.hexer_dir is not configured")
	}
	certPath := filepath.Join(hexerDir, certRelPath)
	if fileExists(certPath) {
		return certPath, nil
	}

	generator := filepath.Join(hexerDir, "scripts", "dev", "generate-dev-certs.sh")
	if !fileExists(generator) {
		return "", fmt.Errorf("no dev certificate at %s and no generator in %s", certPath, hexerDir)
	}
	fmt.Println(ui.Info("no dev certificate yet — generating it with pnpm dev:certs"))

	cmd := exec.Command("pnpm", "dev:certs")
	cmd.Dir = hexerDir
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pnpm dev:certs failed: %w", err)
	}
	if !fileExists(certPath) {
		return "", fmt.Errorf("pnpm dev:certs ran but %s is still missing", certPath)
	}
	return certPath, nil
}

// installCertInto replaces any existing copy rather than adding another, since
// certutil happily stacks duplicates under the same nickname.
func installCertInto(store certStore, certPath string) error {
	removeCertFrom(store)

	out, err := exec.Command("certutil", "-d", "sql:"+store.Dir, "-A",
		"-t", "C,,", "-n", certNickname, "-i", certPath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("certutil import failed: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

// removeCertFrom deletes every copy under the nickname and reports how many
// there were, so repeated imports converge on exactly one.
func removeCertFrom(store certStore) int {
	removed := 0
	for range 20 {
		if !certPresentIn(store) {
			break
		}
		if err := exec.Command("certutil", "-d", "sql:"+store.Dir, "-D", "-n", certNickname).Run(); err != nil {
			break
		}
		removed++
	}
	return removed
}

func certPresentIn(store certStore) bool {
	out, err := exec.Command("certutil", "-d", "sql:"+store.Dir, "-L").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), certNickname)
}

func reportCertStores(stores []certStore) error {
	var rows [][]string
	for _, store := range stores {
		state := "not trusted"
		if certPresentIn(store) {
			state = fmt.Sprintf("trusted (%d cop%s)", countCertIn(store), plural(countCertIn(store)))
		}
		rows = append(rows, []string{store.Label, store.Dir, state})
	}
	fmt.Println(ui.Table([]string{"STORE", "PATH", "STATE"}, rows))
	return nil
}

func countCertIn(store certStore) int {
	out, err := exec.Command("certutil", "-d", "sql:"+store.Dir, "-L").Output()
	if err != nil {
		return 0
	}
	count := 0
	for line := range strings.SplitSeq(string(out), "\n") {
		if strings.Contains(line, certNickname) {
			count++
		}
	}
	return count
}

func plural(n int) string {
	if n == 1 {
		return "y"
	}
	return "ies"
}

// certMacOS cannot do the work itself: the system keychain needs root, so it
// prints the command rather than pretending to have run it.
func certMacOS(cfg *config.Config) error {
	certPath, err := hexerCertPath(cfg)
	if err != nil {
		return err
	}
	fmt.Println(ui.KeyValue("Certificate", certPath))
	if certRemove {
		fmt.Println(ui.Info("remove it with:"))
		fmt.Printf("  sudo security delete-certificate -c '*.acreidentity.dev'\n")
		return nil
	}
	fmt.Println(ui.Info("trust it with:"))
	fmt.Printf("  sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain %q\n", certPath)
	fmt.Println(ui.Info("Firefox keeps its own store: Settings → Certificates → View Certificates → Authorities → Import"))
	return nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func init() {
	certCmd.Flags().BoolVar(&certStatus, "status", false, "report where the certificate is trusted, without changing anything")
	certCmd.Flags().BoolVar(&certRemove, "remove", false, "remove the certificate from every store")
}
