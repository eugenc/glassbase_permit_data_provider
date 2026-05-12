package repair

import (
	"regexp"
	"strings"
)

var hexToken = regexp.MustCompile(`\b([0-9a-fA-F]{7,40})\b`)

// ExtractCommitSHA finds a likely git commit SHA in Claude Code output.
func ExtractCommitSHA(output string) string {
	for _, line := range strings.Split(output, "\n") {
		lower := strings.ToLower(line)
		if !strings.Contains(lower, "commit") {
			continue
		}
		if m := hexToken.FindStringSubmatch(line); len(m) > 1 {
			return m[1]
		}
	}
	return ""
}

// ExtractPRURL finds a GitHub PR URL in output.
func ExtractPRURL(output string) string {
	for _, line := range strings.Split(output, "\n") {
		for _, tok := range strings.Fields(line) {
			if strings.Contains(tok, "github.com") && strings.Contains(tok, "/pull/") {
				return strings.Trim(tok, "()<>,\"'")
			}
		}
	}
	return ""
}

// ParseSuccessFromOutput mirrors heuristics after a non-interactive Claude run.
func ParseSuccessFromOutput(output string, runErr error) bool {
	if runErr != nil {
		return false
	}
	hasSummary := strings.Contains(output, "REPAIR SUMMARY")
	isPaused := strings.Contains(strings.ToLower(output), "marked as paused") ||
		strings.Contains(strings.ToLower(output), "mark county as paused")
	return hasSummary && !isPaused && len(strings.TrimSpace(output)) > 80
}
