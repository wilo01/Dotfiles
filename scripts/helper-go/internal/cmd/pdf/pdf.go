package pdf

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed convert.py
var convertScript []byte

var (
	outParent string
	force     bool
	rawOutput bool
)

// scannedExitCode is emitted by convert.py when the PDF has no usable text layer.
const scannedExitCode = 3

type summary struct {
	Markdown  string `json:"markdown"`
	Pages     int    `json:"pages"`
	Chars     int    `json:"chars"`
	Images    int    `json:"images"`
	TableRows int    `json:"table_rows"`
}

var PdfCmd = &cobra.Command{
	Use:   "pdf <file.pdf>",
	Short: "Convert a PDF to LLM-friendly markdown",
	Long: `Convert a PDF into a folder of LLM-friendly markdown using pymupdf4llm
(via uv). Tables become markdown tables, headings are preserved, and embedded
images are extracted and linked inline.

Report.pdf becomes Report/ containing Report.md and images/.

By default the output is post-processed: <u> tags stripped, ToC dot-leaders
collapsed, <!-- page N --> markers inserted, and YAML front matter prepended.

Examples:
  hlp pdf ~/Downloads/Report.pdf            # -> ~/Downloads/Report/Report.md
  hlp pdf Report.pdf -o ~/Docs              # -> ~/Docs/Report/Report.md
  hlp pdf Report.pdf --raw                  # no post-processing
  hlp pdf Report.pdf --force                # overwrite existing output folder`,
	Args: cobra.ExactArgs(1),
	RunE: runPdf,
}

func init() {
	PdfCmd.Flags().StringVarP(&outParent, "out", "o", "", "parent directory for the output folder (default: the PDF's directory)")
	PdfCmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite an existing output folder")
	PdfCmd.Flags().BoolVar(&rawOutput, "raw", false, "skip post-processing (cleanups, page markers, front matter)")
}

// validateInput checks that path exists, is a regular file, and has a .pdf extension.
func validateInput(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory, expected a PDF file", path)
	}
	if !strings.EqualFold(filepath.Ext(path), ".pdf") {
		return fmt.Errorf("%s does not have a .pdf extension", path)
	}
	return nil
}

// deriveOutput returns the output directory and the PDF basename (without
// extension). The output folder is <parent>/<basename>, where parent defaults
// to the PDF's own directory unless overridden.
func deriveOutput(pdfPath, parentOverride string) (outDir, basename string) {
	basename = strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath))
	parent := filepath.Dir(pdfPath)
	if parentOverride != "" {
		parent = parentOverride
	}
	return filepath.Join(parent, basename), basename
}

func runPdf(cmd *cobra.Command, args []string) error {
	pdfPath, err := filepath.Abs(args[0])
	if err != nil {
		return err
	}
	if err := validateInput(pdfPath); err != nil {
		return err
	}

	uvPath, err := exec.LookPath("uv")
	if err != nil {
		return errors.New("uv is required but not found on PATH\n" +
			"  install it with: curl -LsSf https://astral.sh/uv/install.sh | sh")
	}

	outDir, basename := deriveOutput(pdfPath, outParent)
	if _, err := os.Stat(outDir); err == nil {
		if !force {
			return fmt.Errorf("output folder %s already exists (use --force to overwrite)", outDir)
		}
		if err := os.RemoveAll(outDir); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	scriptFile, err := os.CreateTemp("", "hlp-pdf-*.py")
	if err != nil {
		return err
	}
	defer os.Remove(scriptFile.Name())
	if _, err := scriptFile.Write(convertScript); err != nil {
		return err
	}
	scriptFile.Close()

	convArgs := []string{"run", scriptFile.Name(), pdfPath, outDir, basename}
	if rawOutput {
		convArgs = append(convArgs, "--raw")
	}
	conv := exec.Command(uvPath, convArgs...)
	conv.Stderr = os.Stderr
	stdout, err := conv.Output()
	if err != nil {
		os.RemoveAll(outDir)
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == scannedExitCode {
			return fmt.Errorf("%s appears to be a scanned/image-only PDF — no text layer to convert\n"+
				"  run OCR first:  ocrmypdf %s %s\n"+
				"  then retry with the OCR'd file",
				filepath.Base(pdfPath), pdfPath,
				strings.TrimSuffix(pdfPath, filepath.Ext(pdfPath))+"-ocr.pdf")
		}
		return fmt.Errorf("conversion failed: %w", err)
	}

	// pymupdf4llm may print progress noise on stdout; the summary is the last line.
	var sum summary
	lines := strings.Split(strings.TrimSpace(string(stdout)), "\n")
	line := lines[len(lines)-1]
	if err := json.Unmarshal([]byte(line), &sum); err != nil {
		// Conversion succeeded but the summary was unreadable; still a success.
		fmt.Println(ui.SuccessMsg("Converted " + filepath.Base(pdfPath)))
		fmt.Println(ui.KeyValue("Output", outDir))
		return nil
	}

	fmt.Println(ui.SuccessMsg("Converted " + filepath.Base(pdfPath)))
	fmt.Println(ui.KeyValue("Markdown", sum.Markdown))
	fmt.Println(ui.KeyValue("Pages", fmt.Sprintf("%d", sum.Pages)))
	fmt.Println(ui.KeyValue("Images", fmt.Sprintf("%d", sum.Images)))
	fmt.Println(ui.KeyValue("Table rows", fmt.Sprintf("%d", sum.TableRows)))
	fmt.Println(ui.KeyValue("Text chars", fmt.Sprintf("%d", sum.Chars)))
	return nil
}
