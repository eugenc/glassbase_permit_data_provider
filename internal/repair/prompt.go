package repair

import (
	"fmt"
	"os"

	"github.com/echayko/glassbase_permit_data_provider/internal/registry"
)

// RepairContext bundles what Claude Code needs to repair a county autonomously.
type RepairContext struct {
	County    *registry.CountyConnector
	Trigger   string
	LastError string
	RunsLog   string
}

// BuildPrompt renders the structured Claude Code instruction set. Caller should pass
// LastError / RunsLog already truncated (runner applies EnvLimits truncation before calling).
func BuildPrompt(rc RepairContext, lim EnvLimits) string {
	if lim.CompactPrompt {
		return buildCompactPrompt(rc)
	}
	return buildFullPrompt(rc)
}

func buildFullPrompt(rc RepairContext) string {
	scrapeCMD := resolveCLICmd("scrape-one")
	diagnoseCMD := resolveCLICmd("diagnose")
	onboardCMD := resolveCLICmd("onboard")
	triggerDesc := triggerDescription(rc.Trigger)
	errorSection := ""
	if rc.LastError != "" {
		errorSection = fmt.Sprintf("\nLast recorded error:\n%s\n", rc.LastError)
	}
	runsSection := ""
	if rc.RunsLog != "" {
		runsSection = fmt.Sprintf("\nRecent run history:\n%s\n", rc.RunsLog)
	}

	return fmt.Sprintf(`You are autonomously repairing a broken county permit scraper in the GlassBase permit data provider (module github.com/echayko/glassbase_permit_data_provider).

## Broken county
- County ID:   %s
- County Name: %s
- State:       %s
- URL:         %s
- Source Type: %s
- Trigger:     %s
%s%s

## Repair workflow (in order)

### 1 — Diagnose
Run:
    %s --county=%s

Focus on: selector match count, network calls, body length/SPA.

### 2 — Classify (pick one)
A. Selector drift — CSS no longer matches  
B. SPA timing — content not rendered when fetched  
C. Pagination changed  
D. Source type / API change  
E. HTTP errors (403/429/5xx)  
F. Unknown

### 3 — Fix

**A / C (connector_config only):** Update JSONB in Postgres (psql or migrations policy your team uses). For partial JSON merge, use PostgreSQL jsonb operators. Re-run diagnose after changes.

**B / E (Go code):** Change internal/fetcher (e.g. spa.go, html.go). This needs a normal PR from a dev machine or CI — in a minimal Docker/runtime without a git working tree, describe the exact code change in the REPAIR SUMMARY and mark the county paused if you cannot patch.

**D:** Re-onboard to regenerate config:
    %s --url=%s --county=%s --name=%q --state=%s

### 4 — Verify
    %s --county=%s

Success requires exit 0 and found > 0 (unless the source is genuinely empty — then say so explicitly).

### 5 — Git (only if you have a full clone and credentials)
Connector-only DB updates: optional empty commit for audit.
Go changes: branch + PR; do not merge yourself.

If running without git (e.g. production container): skip git; apply DB fixes with psql against DATABASE_URL and summarize.

### 6 — Report
Print a summary in exactly this block:

REPAIR SUMMARY
==============
County: %s
Root Cause: [A/B/C/D/E/F] — <brief>
Fix Applied: <what changed>
Verification: <scrape-one output line>
Action: <db updated | pr opened at URL | paused | needs human>

`,
		rc.County.CountyID,
		rc.County.CountyName,
		rc.County.State,
		rc.County.URL,
		rc.County.SourceType,
		triggerDesc,
		errorSection,
		runsSection,
		diagnoseCMD,
		rc.County.CountyID,
		onboardCMD,
		rc.County.URL,
		rc.County.CountyID,
		rc.County.CountyName,
		rc.County.State,
		scrapeCMD,
		rc.County.CountyID,
		rc.County.CountyID,
	)
}

func buildCompactPrompt(rc RepairContext) string {
	scrapeCMD := resolveCLICmd("scrape-one")
	diagnoseCMD := resolveCLICmd("diagnose")
	onboardCMD := resolveCLICmd("onboard")
	triggerDesc := triggerDescription(rc.Trigger)
	errorSection := ""
	if rc.LastError != "" {
		errorSection = fmt.Sprintf("\nLast error:\n%s\n", rc.LastError)
	}
	runsSection := ""
	if rc.RunsLog != "" {
		runsSection = fmt.Sprintf("\nRuns:\n%s\n", rc.RunsLog)
	}

	return fmt.Sprintf(`GlassBase permit scraper repair (github.com/echayko/glassbase_permit_data_provider).

County: id=%s name=%s state=%s url=%s source_type=%s trigger=%s%s%s
1) Run %s --county=%s (selectors, network, body/SPA).
2) Classify: A drift B SPA C pagination D API/HTML change E HTTP F unknown.
3) Fix: A/C update connector JSONB in Postgres; B/E patch internal/fetcher (or describe + pause if no git); D re-onboard: %s --url=%s --county=%s --name=%q --state=%s
4) Verify: %s --county=%s (exit 0; found>0 unless truly empty).
5) Git: branch+PR only if you have a clone; else psql + summarize.
6) Output this block:

REPAIR SUMMARY
==============
County: %s
Root Cause: [A/B/C/D/E/F] — <brief>
Fix Applied: <what changed>
Verification: <scrape-one line>
Action: <db | PR URL | paused | human>

`,
		rc.County.CountyID,
		rc.County.CountyName,
		rc.County.State,
		rc.County.URL,
		rc.County.SourceType,
		triggerDesc,
		errorSection,
		runsSection,
		diagnoseCMD,
		rc.County.CountyID,
		onboardCMD,
		rc.County.URL,
		rc.County.CountyID,
		rc.County.CountyName,
		rc.County.State,
		scrapeCMD,
		rc.County.CountyID,
		rc.County.CountyID,
	)
}

func resolveCLICmd(name string) string {
	if os.Getenv("GLASSBASE_CLI_WRAPPER") == "binary" {
		return "./" + name
	}
	return "go run ./cmd/" + name
}

func triggerDescription(trigger string) string {
	switch trigger {
	case "zero_records":
		return "Scrape succeeded but returned 0 records (likely selector or pagination)"
	case "health_probe":
		return "Health probe reported an unhealthy county endpoint"
	case "manual":
		return "Manual repair triggered from admin or ops"
	default:
		return trigger
	}
}
