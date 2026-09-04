package agent

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// newNSSDB creates an empty NSS database, the store certutil operates on.
func newNSSDB(t *testing.T) certStore {
	t.Helper()
	if _, err := exec.LookPath("certutil"); err != nil {
		t.Skip("certutil not installed")
	}
	dir := t.TempDir()
	out, err := exec.Command("certutil", "-d", "sql:"+dir, "-N", "--empty-password").CombinedOutput()
	if err != nil {
		t.Skipf("cannot create an NSS db here: %s", out)
	}
	return certStore{Label: "test", Dir: dir}
}

// selfSignedCA writes a CA certificate the importer can install.
func selfSignedCA(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	certPath := filepath.Join(dir, "acreidentity-dev.crt")
	keyPath := filepath.Join(dir, "acreidentity-dev.key")

	out, err := exec.Command("openssl", "req", "-x509", "-newkey", "rsa:2048", "-sha256",
		"-days", "1", "-nodes", "-keyout", keyPath, "-out", certPath,
		"-subj", "/CN=*.acreidentity.dev",
		"-addext", "basicConstraints=critical,CA:TRUE",
		"-addext", "subjectAltName=DNS:*.acrid.dev,DNS:acrid.dev").CombinedOutput()
	if err != nil {
		t.Skipf("openssl unavailable: %s", out)
	}
	return certPath
}

func TestInstallCertIsIdempotent(t *testing.T) {
	store := newNSSDB(t)
	cert := selfSignedCA(t)

	if certPresentIn(store) {
		t.Fatal("a fresh store already contains the certificate")
	}
	for range 3 {
		if err := installCertInto(store, cert); err != nil {
			t.Fatalf("installCertInto: %v", err)
		}
	}

	if !certPresentIn(store) {
		t.Fatal("certificate not present after install")
	}
	// Repeated imports must converge on one entry: certutil stacks duplicates
	// under the same nickname, which is how five copies accumulated by hand.
	if got := countCertIn(store); got != 1 {
		t.Errorf("store holds %d copies, want exactly 1", got)
	}
}

func TestRemoveCertClearsEveryCopy(t *testing.T) {
	store := newNSSDB(t)

	// Duplicates stack when the cert is regenerated: certutil collapses an
	// identical import, but each `pnpm dev:certs` run produces a new key, so the
	// same nickname accumulates distinct certificates.
	for range 3 {
		exec.Command("certutil", "-d", "sql:"+store.Dir, "-A",
			"-t", "C,,", "-n", certNickname, "-i", selfSignedCA(t)).Run()
	}
	if got := countCertIn(store); got < 2 {
		t.Fatalf("expected stacked duplicates, got %d", got)
	}

	removed := removeCertFrom(store)
	if removed < 2 {
		t.Errorf("removed %d copies, expected every duplicate", removed)
	}
	if certPresentIn(store) {
		t.Error("certificate still present after removal")
	}
}

func TestRemoveCertOnAnEmptyStoreIsANoOp(t *testing.T) {
	store := newNSSDB(t)

	if got := removeCertFrom(store); got != 0 {
		t.Errorf("removed %d from an empty store", got)
	}
}

func TestDiscoverCertStoresFindsNothingInAnEmptyHome(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if got := discoverCertStores(); len(got) != 0 {
		t.Errorf("expected no stores, got %+v", got)
	}
}

// Chrome's shared store and every Firefox profile are separate databases, and
// the OS trust commands cover neither.
func TestDiscoverCertStoresFindsChromeAndFirefoxProfiles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := os.MkdirAll(filepath.Join(home, ".pki", "nssdb"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, profile := range []string{"abc.default-release", "xyz.dev-edition"} {
		dir := filepath.Join(home, ".mozilla", "firefox", profile)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "cert9.db"), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	stores := discoverCertStores()
	if len(stores) != 3 {
		t.Fatalf("expected Chrome plus two Firefox profiles, got %+v", stores)
	}
	if stores[0].Label != "Chrome/Chromium" {
		t.Errorf("shared store should come first, got %q", stores[0].Label)
	}
}

func TestHexerCertPathErrorsWhenUnconfigured(t *testing.T) {
	cfg := workspace(t)
	cfg.Agent.Hexer.HexerDir = ""

	if _, err := hexerCertPath(cfg); err == nil {
		t.Fatal("expected an error with no hexer_dir configured")
	}
}

func TestHexerCertPathFindsAnExistingCert(t *testing.T) {
	cfg := workspace(t)
	hexerDir := t.TempDir()
	cfg.Agent.Hexer.HexerDir = hexerDir

	certPath := filepath.Join(hexerDir, certRelPath)
	if err := os.MkdirAll(filepath.Dir(certPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(certPath, []byte("-----BEGIN CERTIFICATE-----\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := hexerCertPath(cfg)
	if err != nil {
		t.Fatalf("hexerCertPath: %v", err)
	}
	if got != certPath {
		t.Errorf("got %q, want %q", got, certPath)
	}
}

func TestPlural(t *testing.T) {
	if plural(1) != "y" || plural(0) != "ies" || plural(5) != "ies" {
		t.Error("copy/copies wording is wrong")
	}
}
