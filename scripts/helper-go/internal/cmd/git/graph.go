package git

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dariuszw/hlp/internal/ui"
)

const (
	graphNodeChar   = "●"
	graphLineChar   = "─"
	graphBranchChar = "│"
	graphMergeChar  = "└"
	graphForkChar   = "├"
)

// Table column widths (accounting for Unicode characters)
const (
	colGraphWidth   = 8
	colMessageWidth = 42
	colDateWidth    = 14
	colAuthorWidth  = 13
	colCommitWidth  = 10
)

// printGraph renders an ASCII visualization of the branch divergence
func printGraph(r *CompareResult) {
	// Calculate graph dimensions
	maxBehind := r.Behind
	maxAhead := r.Ahead
	if maxBehind == 0 {
		maxBehind = 1
	}
	if maxAhead == 0 {
		maxAhead = 1
	}

	// Limit node display for readability
	displayBehind := min(maxBehind, 5)
	displayAhead := min(maxAhead, 5)

	// Build the branch line (what's on branch but not on commit)
	branchLine := buildBranchLine(r.Branch, r.BranchHead, displayBehind, r.Behind)
	fmt.Println(branchLine)

	// Build the divergence point
	divergeLine := buildDivergeLine(displayBehind, r.MergeBase)
	fmt.Println(divergeLine)

	// Build the commit line (what's on commit but not on branch)
	commitLine := buildCommitLine(r.CommitShort, displayAhead, r.Ahead, displayBehind)
	fmt.Println(commitLine)
}

func buildBranchLine(branchName, branchHead string, displayBehind, totalBehind int) string {
	var sb strings.Builder

	// Branch label
	label := fmt.Sprintf("%s ", branchName)
	sb.WriteString(ui.Muted.Render(label))

	// Line to merge base
	sb.WriteString(ui.Muted.Render(strings.Repeat(graphLineChar, 2)))

	// Merge base node
	sb.WriteString(ui.Primary.Render(graphNodeChar))

	// Commits behind (on branch, not on commit)
	for i := 0; i < displayBehind; i++ {
		sb.WriteString(ui.WarningText.Render(graphLineChar + graphLineChar + graphNodeChar))
	}

	if totalBehind > displayBehind {
		sb.WriteString(ui.Muted.Render("···"))
	}

	// Branch head indicator
	sb.WriteString(ui.Muted.Render(graphLineChar + graphLineChar))
	sb.WriteString(ui.WarningText.Render(fmt.Sprintf(" (%s)", branchHead)))

	return sb.String()
}

func buildDivergeLine(displayBehind int, mergeBase string) string {
	var sb strings.Builder

	// Padding for branch label alignment
	sb.WriteString(strings.Repeat(" ", 8))

	// Vertical line at divergence point
	sb.WriteString(ui.Primary.Render(graphBranchChar))

	// Show merge base
	sb.WriteString(ui.Muted.Render(fmt.Sprintf(" merge-base: %s", mergeBase)))

	return sb.String()
}

func buildCommitLine(commitRef string, displayAhead, totalAhead, displayBehind int) string {
	var sb strings.Builder

	// Commit label with padding
	label := fmt.Sprintf("%s ", commitRef)
	sb.WriteString(ui.Value.Render(label))

	// Padding to align with merge base
	padLen := 8 - len(label)
	if padLen > 0 {
		sb.WriteString(strings.Repeat(" ", padLen))
	}

	// Fork from merge base
	sb.WriteString(ui.Primary.Render(graphMergeChar))

	// Commits ahead (on commit, not on branch)
	for i := 0; i < displayAhead; i++ {
		sb.WriteString(ui.Success.Render(graphLineChar + graphLineChar + graphNodeChar))
	}

	if totalAhead > displayAhead {
		sb.WriteString(ui.Muted.Render("···"))
	}

	// Status indicator
	if totalAhead > 0 {
		sb.WriteString(ui.Success.Render(fmt.Sprintf(" (+%d ahead)", totalAhead)))
	} else {
		sb.WriteString(ui.Muted.Render(" (up to date)"))
	}

	return sb.String()
}

// ============================================================================
// Table-based visualization (VSCode Git Graph style)
// ============================================================================

// printTable renders commits in a table format like VSCode Git Graph
func printTable(r *CompareResult, maxCommits int) {
	// Table styles
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorPrimary)
	borderStyle := lipgloss.NewStyle().Foreground(ui.ColorMuted)
	targetStyle := lipgloss.NewStyle().Foreground(ui.ColorHighlight).Bold(true)

	// Print header
	printTableBorder(borderStyle, "┌", "┬", "┐")
	printTableHeader(headerStyle, borderStyle)
	printTableBorder(borderStyle, "├", "┼", "┤")

	// Collect all commits for display
	// First: commits on branch (behind) - these are newer, show at top
	behindCount := 0
	for i, c := range r.BehindEnhanced {
		if i >= maxCommits {
			break
		}
		graphSymbol := graphBranchChar
		if i == 0 {
			graphSymbol = graphNodeChar
		}
		branchLabel := ""
		if i == 0 {
			branchLabel = r.Branch
		}
		printTableRow(c, graphSymbol, branchLabel, false, 0, borderStyle)
		behindCount++
	}

	// Show "more" indicator for behind commits
	if r.Behind > maxCommits {
		remaining := r.Behind - maxCommits
		printTableMoreRow(remaining, "on "+r.Branch, borderStyle)
	}

	// Mark where target commit diverges
	if behindCount > 0 || r.Behind > 0 {
		// Print the target commit with fork indicator
		if len(r.AheadEnhanced) > 0 {
			c := r.AheadEnhanced[0]
			printTableRow(c, graphForkChar+graphNodeChar, "", true, 1, borderStyle)
			// Print remaining ahead commits
			for i := 1; i < len(r.AheadEnhanced) && i < maxCommits; i++ {
				printTableRow(r.AheadEnhanced[i], graphBranchChar, "", false, 1, borderStyle)
			}
		} else {
			// Target has no unique commits, show merge point
			printTableMergeRow(r.CommitShort, r.MergeBase, targetStyle, borderStyle)
		}
	} else {
		// No behind commits, just show ahead commits
		for i, c := range r.AheadEnhanced {
			if i >= maxCommits {
				break
			}
			graphSymbol := graphNodeChar
			if i > 0 {
				graphSymbol = graphBranchChar
			}
			printTableRow(c, graphSymbol, "", i == 0, 1, borderStyle)
		}
	}

	// Show "more" indicator for ahead commits
	if r.Ahead > maxCommits {
		remaining := r.Ahead - maxCommits
		printTableMoreRow(remaining, "ahead", borderStyle)
	}

	// Print footer
	printTableBorder(borderStyle, "└", "┴", "┘")
}

func printTableBorder(style lipgloss.Style, left, mid, right string) {
	fmt.Printf("%s%s%s%s%s%s%s%s%s%s%s\n",
		style.Render(left),
		style.Render(strings.Repeat("─", colGraphWidth)),
		style.Render(mid),
		style.Render(strings.Repeat("─", colMessageWidth)),
		style.Render(mid),
		style.Render(strings.Repeat("─", colDateWidth)),
		style.Render(mid),
		style.Render(strings.Repeat("─", colAuthorWidth)),
		style.Render(mid),
		style.Render(strings.Repeat("─", colCommitWidth)),
		style.Render(right),
	)
}

func printTableHeader(headerStyle, borderStyle lipgloss.Style) {
	fmt.Printf("%s %s %s %s %s %s %s %s %s %s %s\n",
		borderStyle.Render("│"),
		headerStyle.Render(padRight("Graph", colGraphWidth-2)),
		borderStyle.Render("│"),
		headerStyle.Render(padRight("Description", colMessageWidth-2)),
		borderStyle.Render("│"),
		headerStyle.Render(padRight("Date", colDateWidth-2)),
		borderStyle.Render("│"),
		headerStyle.Render(padRight("Author", colAuthorWidth-2)),
		borderStyle.Render("│"),
		headerStyle.Render(padRight("Commit", colCommitWidth-2)),
		borderStyle.Render("│"),
	)
}

func printTableRow(c EnhancedCommitInfo, graphSymbol, branchLabel string, isTarget bool, colorIdx int, borderStyle lipgloss.Style) {
	branchStyle := ui.BranchStyle(colorIdx)
	targetMarker := ""
	msgStyle := lipgloss.NewStyle()

	if isTarget {
		targetMarker = "◀ "
		msgStyle = lipgloss.NewStyle().Foreground(ui.ColorHighlight)
	}

	// Build message with optional branch label
	msg := c.Message
	if branchLabel != "" {
		msg = fmt.Sprintf("[%s] %s", branchLabel, c.Message)
	}
	msg = targetMarker + msg

	// Format author name
	author := formatAuthorShort(c.Author)

	fmt.Printf("%s %s %s %s %s %s %s %s %s %s %s\n",
		borderStyle.Render("│"),
		branchStyle.Render(padRight(graphSymbol, colGraphWidth-2)),
		borderStyle.Render("│"),
		msgStyle.Render(truncateRight(msg, colMessageWidth-2)),
		borderStyle.Render("│"),
		ui.Muted.Render(padRight(c.Date, colDateWidth-2)),
		borderStyle.Render("│"),
		ui.Muted.Render(truncateRight(author, colAuthorWidth-2)),
		borderStyle.Render("│"),
		ui.Secondary.Render(padRight(c.ShortHash, colCommitWidth-2)),
		borderStyle.Render("│"),
	)
}

func printTableMergeRow(commitShort, mergeBase string, targetStyle, borderStyle lipgloss.Style) {
	msg := fmt.Sprintf("◀ merge-base: %s", mergeBase)
	fmt.Printf("%s %s %s %s %s %s %s %s %s %s %s\n",
		borderStyle.Render("│"),
		ui.Primary.Render(padRight(graphMergeChar, colGraphWidth-2)),
		borderStyle.Render("│"),
		targetStyle.Render(truncateRight(msg, colMessageWidth-2)),
		borderStyle.Render("│"),
		ui.Muted.Render(padRight("", colDateWidth-2)),
		borderStyle.Render("│"),
		ui.Muted.Render(padRight("", colAuthorWidth-2)),
		borderStyle.Render("│"),
		ui.Secondary.Render(padRight(commitShort, colCommitWidth-2)),
		borderStyle.Render("│"),
	)
}

func printTableMoreRow(count int, label string, borderStyle lipgloss.Style) {
	msg := fmt.Sprintf("... and %d more %s", count, label)
	fmt.Printf("%s %s %s %s %s %s %s %s %s %s %s\n",
		borderStyle.Render("│"),
		ui.Muted.Render(padRight("", colGraphWidth-2)),
		borderStyle.Render("│"),
		ui.Muted.Render(truncateRight(msg, colMessageWidth-2)),
		borderStyle.Render("│"),
		ui.Muted.Render(padRight("", colDateWidth-2)),
		borderStyle.Render("│"),
		ui.Muted.Render(padRight("", colAuthorWidth-2)),
		borderStyle.Render("│"),
		ui.Muted.Render(padRight("", colCommitWidth-2)),
		borderStyle.Render("│"),
	)
}

// Helper functions
func padRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

func truncateRight(s string, width int) string {
	if len(s) <= width {
		return padRight(s, width)
	}
	if width <= 3 {
		return s[:width]
	}
	return s[:width-3] + "..."
}

func formatAuthorShort(name string) string {
	parts := strings.Fields(name)
	if len(parts) >= 2 {
		return fmt.Sprintf("%s. %s", string(parts[0][0]), parts[len(parts)-1])
	}
	return name
}
