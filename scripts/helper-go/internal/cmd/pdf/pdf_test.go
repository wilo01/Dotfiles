package pdf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDeriveOutput_DefaultNextToPDF(t *testing.T) {
	outDir, basename := deriveOutput("/home/user/Downloads/Report.pdf", "")
	if outDir != "/home/user/Downloads/Report" {
		t.Errorf("outDir = %q, want /home/user/Downloads/Report", outDir)
	}
	if basename != "Report" {
		t.Errorf("basename = %q, want Report", basename)
	}
}

func TestDeriveOutput_ParentOverride(t *testing.T) {
	outDir, basename := deriveOutput("/home/user/Downloads/Report.pdf", "/home/user/Docs")
	if outDir != "/home/user/Docs/Report" {
		t.Errorf("outDir = %q, want /home/user/Docs/Report", outDir)
	}
	if basename != "Report" {
		t.Errorf("basename = %q, want Report", basename)
	}
}

func TestDeriveOutput_DottedName(t *testing.T) {
	outDir, basename := deriveOutput("/tmp/my.report.v2.pdf", "")
	if basename != "my.report.v2" {
		t.Errorf("basename = %q, want my.report.v2", basename)
	}
	if outDir != "/tmp/my.report.v2" {
		t.Errorf("outDir = %q, want /tmp/my.report.v2", outDir)
	}
}

func TestValidateInput_Missing(t *testing.T) {
	if err := validateInput(filepath.Join(t.TempDir(), "nope.pdf")); err == nil {
		t.Error("expected error for missing file")
	}
}

func TestValidateInput_Directory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "folder.pdf")
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := validateInput(dir); err == nil {
		t.Error("expected error for directory input")
	}
}

func TestValidateInput_WrongExtension(t *testing.T) {
	file := filepath.Join(t.TempDir(), "doc.txt")
	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := validateInput(file); err == nil {
		t.Error("expected error for non-pdf extension")
	}
}

func TestValidateInput_UppercaseExtension(t *testing.T) {
	file := filepath.Join(t.TempDir(), "doc.PDF")
	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := validateInput(file); err != nil {
		t.Errorf("expected .PDF to be accepted, got %v", err)
	}
}
