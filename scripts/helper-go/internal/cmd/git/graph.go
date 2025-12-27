package git

import (
	"fmt"
	"strings"

	"github.com/dariuszw/hlp/internal/ui"
)

const (
	graphNodeChar   = "●"
	graphLineChar   = "─"
	graphBranchChar = "│"
	graphMergeChar  = "└"
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
