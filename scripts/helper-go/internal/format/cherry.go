package format

import (
	"fmt"
	"regexp"
	"strings"
)

// versionSuffixPattern matches version suffixes like "-13.2av", "-13av", "13.2AV" at end of string.
// The leading dash is optional to handle both "branch-13.2av" and edge cases.
var versionSuffixPattern = regexp.MustCompile(`(?i)-(\d+(?:\.\d+)?av)$`)

// maintenanceVersionPattern extracts version from "maintenance/13.1AV" format.
var maintenanceVersionPattern = regexp.MustCompile(`(?i)(?:maintenance/)(\d+(?:\.\d+)?av)$`)

// DeriveCherryBranch creates the new branch name for a cherry-pick.
//
// Two cases:
//  1. Source has version suffix → replace it:
//     "vis-6893-...-13.2av" + "maintenance/13.1AV" → "vis-6893-...-13.1av"
//  2. Source has NO version suffix → append it:
//     "VIS-6457-...-keyboard-default" + "maintenance/13.1AV" → "VIS-6457-...-keyboard-default-13.1av"
func DeriveCherryBranch(sourceBranch, targetBranch string) (string, error) {
	targetVersion, err := extractTargetVersion(targetBranch)
	if err != nil {
		return "", err
	}
	lowerVersion := strings.ToLower(targetVersion)

	if versionSuffixPattern.MatchString(sourceBranch) {
		// Replace existing version suffix
		result := versionSuffixPattern.ReplaceAllString(sourceBranch, "-"+lowerVersion)
		return result, nil
	}

	// No version suffix — append
	return sourceBranch + "-" + lowerVersion, nil
}

// ExtractMaintenanceVersion extracts the version string from a maintenance branch name.
// e.g., "maintenance/13.1AV" → "13.1AV", "maintenance/12AV" → "12AV"
// Returns empty string if the branch doesn't match the expected pattern.
func ExtractMaintenanceVersion(branch string) string {
	match := maintenanceVersionPattern.FindStringSubmatch(branch)
	if len(match) >= 2 {
		return match[1]
	}
	return ""
}

// extractTargetVersion extracts and validates the version from a target branch.
func extractTargetVersion(targetBranch string) (string, error) {
	match := maintenanceVersionPattern.FindStringSubmatch(targetBranch)
	if len(match) < 2 {
		return "", fmt.Errorf("cannot extract version from target branch %q (expected maintenance/X.YAV)", targetBranch)
	}
	return match[1], nil
}
