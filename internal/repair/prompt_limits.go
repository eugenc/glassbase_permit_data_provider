package repair

import (
	"os"
	"strconv"
	"strings"
)

// EnvLimits configures repair prompt sizing and pacing (Anthropic TPM / Claude Code budgets).
type EnvLimits struct {
	MaxLastErrRunes int
	MaxRunsLogRunes int
	MaxRunErrRunes  int
	CronStaggerSecs int // sleep between cron-driven counties (>0 skips first iteration)
	CompactPrompt   bool
}

// LimitsFromEnv returns limits with sane defaults. Token budgets are coarse; TPM also includes
// Claude Code’s own attachments and tool context beyond this prompt slice.
func LimitsFromEnv() EnvLimits {
	return EnvLimits{
		MaxLastErrRunes: getenvIntClamp("GLASSBASE_REPAIR_MAX_LAST_ERROR_RUNES", 2400, 200, 32_000),
		MaxRunsLogRunes: getenvIntClamp("GLASSBASE_REPAIR_MAX_RUNS_LOG_RUNES", 4800, 200, 32_000),
		MaxRunErrRunes:  getenvIntClamp("GLASSBASE_REPAIR_MAX_RUN_ERR_RUNES", 600, 50, 4000),
		CronStaggerSecs: getenvIntClamp("GLASSBASE_REPAIR_CRON_STAGGER_SECONDS", 0, 0, 600),
		CompactPrompt:   getenvBoolTruthy("GLASSBASE_REPAIR_COMPACT_PROMPT"),
	}
}

func getenvBoolTruthy(key string) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	return v == "1" || v == "true" || v == "yes"
}

func getenvIntClamp(key string, def, minv, maxv int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	if n < minv {
		return minv
	}
	if n > maxv {
		return maxv
	}
	return n
}

func truncateUTF8Runes(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return s
	}
	s = strings.TrimSpace(s)
	n := 0
	for i := range s {
		if n >= maxRunes {
			return strings.TrimRight(s[:i], " \t\r\n") + "\n… [truncated for prompt size]"
		}
		n++
	}
	return s
}
