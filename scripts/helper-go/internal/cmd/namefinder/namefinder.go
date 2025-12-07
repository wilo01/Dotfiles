package namefinder

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/dariuszw/hlp/internal/checker"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

var (
	workers      int
	outputFormat string
	outputFile   string
	skipDomains  bool
	skipPlatform bool
	tlds         []string
)

// NameFinderCmd is the name-finder command
var NameFinderCmd = &cobra.Command{
	Use:   "name-finder <names...>",
	Short: "Check name availability across domains and platforms",
	Long: `Check if a name is available on domains and developer platforms.

Checks:
- Domains: .com, .net, .org, .io, .app, .dev, .pl, etc.
- Platforms: GitHub, npm, PyPI, Docker Hub, Crates.io

Examples:
  hlp name-finder myproject
  hlp name-finder myproject coolname bestlib
  hlp name-finder myproject --tlds com,io,dev
  hlp name-finder myproject --output json --file results.json`,
	Args: cobra.MinimumNArgs(1),
	Run:  runNameFinder,
}

func init() {
	NameFinderCmd.Flags().IntVarP(&workers, "workers", "w", 10, "number of parallel workers")
	NameFinderCmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "output format: table, json, csv, markdown")
	NameFinderCmd.Flags().StringVarP(&outputFile, "file", "f", "", "output file path")
	NameFinderCmd.Flags().BoolVar(&skipDomains, "skip-domains", false, "skip domain checks")
	NameFinderCmd.Flags().BoolVar(&skipPlatform, "skip-platforms", false, "skip platform checks")
	NameFinderCmd.Flags().StringSliceVar(&tlds, "tlds", nil, "TLDs to check (default: com,net,org,io,app,dev,pl)")
}

func runNameFinder(cmd *cobra.Command, args []string) {
	names := args

	fmt.Println(ui.Header("Name Finder"))
	fmt.Println()
	fmt.Printf("Checking %d name(s) with %d workers...\n\n", len(names), workers)

	// Build checkers
	var checkers []checker.Checker

	if !skipDomains {
		tldsToCheck := tlds
		if len(tldsToCheck) == 0 {
			tldsToCheck = checker.CommonTLDs()
		}
		checkers = append(checkers, checker.CreateDomainCheckers(tldsToCheck)...)
	}

	if !skipPlatform {
		checkers = append(checkers, checker.CreatePlatformCheckers()...)
	}

	if len(checkers) == 0 {
		fmt.Println(ui.Error("No checkers enabled. Remove --skip flags."))
		return
	}

	// Run checks
	executor := checker.NewExecutor(workers)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	start := time.Now()
	results := executor.Execute(ctx, names, checkers)
	elapsed := time.Since(start)

	// Output results
	switch outputFormat {
	case "json":
		outputJSON(results, outputFile)
	case "csv":
		outputCSV(results, outputFile)
	case "markdown", "md":
		outputMarkdown(results, outputFile)
	default:
		outputTable(results)
	}

	fmt.Printf("\nCompleted in %s\n", elapsed.Round(time.Millisecond))
}

func outputTable(results map[string]map[string]checker.Result) {
	for name, services := range results {
		fmt.Println(ui.Title.Render(name))
		fmt.Println()

		// Group by category
		var domains, platforms []checker.Result
		for _, result := range services {
			if strings.HasPrefix(result.Service, "domain.") {
				domains = append(domains, result)
			} else {
				platforms = append(platforms, result)
			}
		}

		// Sort by service name
		sort.Slice(domains, func(i, j int) bool {
			return domains[i].Service < domains[j].Service
		})
		sort.Slice(platforms, func(i, j int) bool {
			return platforms[i].Service < platforms[j].Service
		})

		// Print domains
		if len(domains) > 0 {
			fmt.Println(ui.Subtitle.Render("  Domains"))
			for _, r := range domains {
				tld := strings.TrimPrefix(r.Service, "domain.")
				domain := name + "." + tld
				icon := statusIcon(r.Status)
				fmt.Printf("    %s %s\n", icon, domain)
			}
			fmt.Println()
		}

		// Print platforms
		if len(platforms) > 0 {
			fmt.Println(ui.Subtitle.Render("  Platforms"))
			for _, r := range platforms {
				icon := statusIcon(r.Status)
				fmt.Printf("    %s %s\n", icon, r.Service)
			}
			fmt.Println()
		}

		// Summary
		available := 0
		taken := 0
		for _, r := range services {
			switch r.Status {
			case checker.StatusAvailable:
				available++
			case checker.StatusTaken:
				taken++
			}
		}
		fmt.Printf("  %s available, %s taken\n\n",
			ui.Success.Render(fmt.Sprintf("%d", available)),
			ui.ErrorText.Render(fmt.Sprintf("%d", taken)))
	}
}

func statusIcon(status checker.Status) string {
	switch status {
	case checker.StatusAvailable:
		return ui.Success.Render("✓")
	case checker.StatusTaken:
		return ui.ErrorText.Render("✗")
	case checker.StatusRateLimited:
		return ui.WarningText.Render("⚠")
	default:
		return ui.Muted.Render("?")
	}
}

func outputJSON(results map[string]map[string]checker.Result, file string) {
	data, _ := json.MarshalIndent(results, "", "  ")
	if file != "" {
		os.WriteFile(file, data, 0644)
		fmt.Printf("Results written to %s\n", file)
	} else {
		fmt.Println(string(data))
	}
}

func outputCSV(results map[string]map[string]checker.Result, file string) {
	var output *csv.Writer
	if file != "" {
		f, err := os.Create(file)
		if err != nil {
			fmt.Println(ui.Error("Failed to create file: " + err.Error()))
			return
		}
		defer f.Close()
		output = csv.NewWriter(f)
	} else {
		output = csv.NewWriter(os.Stdout)
	}
	defer output.Flush()

	// Header
	output.Write([]string{"name", "service", "status", "confidence"})

	// Data
	for name, services := range results {
		for _, r := range services {
			output.Write([]string{
				name,
				r.Service,
				string(r.Status),
				fmt.Sprintf("%.2f", r.Confidence),
			})
		}
	}

	if file != "" {
		fmt.Printf("Results written to %s\n", file)
	}
}

func outputMarkdown(results map[string]map[string]checker.Result, file string) {
	var sb strings.Builder

	for name, services := range results {
		sb.WriteString(fmt.Sprintf("## %s\n\n", name))

		sb.WriteString("| Service | Status |\n")
		sb.WriteString("|---------|--------|\n")

		for _, r := range services {
			status := string(r.Status)
			if r.Status == checker.StatusAvailable {
				status = "✅ " + status
			} else if r.Status == checker.StatusTaken {
				status = "❌ " + status
			}
			sb.WriteString(fmt.Sprintf("| %s | %s |\n", r.Service, status))
		}
		sb.WriteString("\n")
	}

	if file != "" {
		os.WriteFile(file, []byte(sb.String()), 0644)
		fmt.Printf("Results written to %s\n", file)
	} else {
		fmt.Print(sb.String())
	}
}
