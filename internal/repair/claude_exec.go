package repair

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func repairStdBufEnabled() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("GLASSBASE_REPAIR_STDBUF")))
	if v != "" && (v == "0" || v == "false" || v == "off") {
		return false
	}
	return true
}

// maybeLineBuffered runs program under GNU stdbuf (line-buffered stdout/stderr) when available so
// the admin SSE stream receives incremental Claude Code output instead of one blob at EOF.
func maybeLineBuffered(ctx context.Context, program string, args []string) *exec.Cmd {
	if program == "" {
		return exec.CommandContext(ctx, program, args...)
	}
	if !repairStdBufEnabled() {
		return exec.CommandContext(ctx, program, args...)
	}
	stdbufPath, err := exec.LookPath("stdbuf")
	if err != nil || strings.TrimSpace(stdbufPath) == "" {
		return exec.CommandContext(ctx, program, args...)
	}
	wrapped := append([]string{"-oL", "-eL", program}, args...)
	log.Printf("[repair-ai] Claude Code subprocess wrapped with %s (-oL -eL)", stdbufPath)
	return exec.CommandContext(ctx, stdbufPath, wrapped...)
}

func claudeArgsForRepair(prompt string) []string {
	// Non-interactive print mode args. Claude Code refuses --dangerously-skip-permissions when running
	// as root (common for Docker/Railway); omit it on EUID 0 so repair can still start.
	args := []string{"--print"}
	if os.Geteuid() != 0 {
		args = append(args, "--dangerously-skip-permissions")
	} else {
		log.Printf("[repair-ai] Claude Code: omitting --dangerously-skip-permissions (process is root; Claude Code forbids that flag under root)")
	}
	args = append(args, prompt)
	return args
}

// claudeRepairCommand builds exec.Cmd for non-interactive Claude Code with the given prompt.
// Resolution: CLAUDE_BIN → claude on PATH → npx -y @anthropic-ai/claude-code (if npx exists or CLAUDE_USE_NPX=1).
func claudeRepairCommand(ctx context.Context, prompt string) (*exec.Cmd, error) {
	claudeArgs := claudeArgsForRepair(prompt)

	if bin := strings.TrimSpace(os.Getenv("CLAUDE_BIN")); bin != "" {
		log.Printf("[repair-ai] Claude Code CLI resolved to CLAUDE_BIN=%q", bin)
		return maybeLineBuffered(ctx, bin, claudeArgs), nil
	}

	forceNpx := strings.EqualFold(strings.TrimSpace(os.Getenv("CLAUDE_USE_NPX")), "1") ||
		strings.EqualFold(strings.TrimSpace(os.Getenv("CLAUDE_USE_NPX")), "true")

	if !forceNpx {
		if path, err := exec.LookPath("claude"); err == nil && path != "" {
			log.Printf("[repair-ai] Claude Code CLI resolved to PATH claude=%q", path)
			return maybeLineBuffered(ctx, path, claudeArgs), nil
		}
	}

	npxPath, err := exec.LookPath("npx")
	if err == nil && npxPath != "" {
		if forceNpx {
			log.Printf("[repair-ai] Claude Code CLI: npx @anthropic-ai/claude-code (CLAUDE_USE_NPX forced)")
		} else {
			log.Printf("[repair-ai] Claude Code CLI: claude not on PATH; using npx @anthropic-ai/claude-code (first run may install packages)")
		}
		args := append([]string{"-y", "@anthropic-ai/claude-code"}, claudeArgs...)
		return maybeLineBuffered(ctx, npxPath, args), nil
	}
	return nil, fmt.Errorf(
		"Claude Code CLI not found (looked for \"claude\" on PATH and \"npx\"). " +
			"Install globally: npm install -g @anthropic-ai/claude-code, " +
			"or set CLAUDE_BIN to your claude executable, " +
			"or set CLAUDE_USE_NPX=1 with Node/npm installed",
	)
}
