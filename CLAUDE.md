# GlassBase permit data provider — Claude Code project context

## What this project does

This service scrapes county building permit pages on a weekly cron. Each county has a `connector_config` JSON blob in Postgres describing how the scraper extracts permits for that jurisdiction.

Module path: `github.com/echayko/glassbase_permit_data_provider`.

## Architecture

- **Fetcher** (`internal/fetcher/`) — HTML, SPA (Rod / headless Chrome), or API requests
- **Generator** (`internal/generator/`) — builds connector configs via Claude API from page content
- **Scraper** (`internal/scraper/`) — parses pages using connector config, upserts into `permits_{county_id}` tables
- **Runner** (`internal/runner/`) — cron + worker pool for batch scrapes and single-county runs
- **Monitor** (`internal/monitor/`) — health probe, consecutive zero-record detection, repair orchestration
- **Repair (Claude Code)** (`internal/repair/`) — shells out to the `claude` CLI with structured prompts
- **API** (`internal/api/`) — admin REST API
- **Frontend** (`frontend/`) — Vite + React admin UI (build output under `internal/static/web`)

## Database

- `county_connectors` — one row per county (`connector_config` JSONB, `status`)
- `scrape_runs` — every scrape (`status`, `records_found`, `records_inserted`, `error_message`)
- `repair_runs` — Claude Code autonomous repair attempts
- `permits_{county_id}` — per-county permit tables (`raw_data` JSONB plus dynamic extraction columns)

Postgres URL: use **`DATABASE_URL_DEV`** and **`DATABASE_URL_PROD`**, or a single **`DATABASE_URL`** (Railway / CI). Resolution is in `config.PickDatabaseURL` (see `.env.example`).

## Common failure modes

1. **Selector drift** — CSS selectors in `extraction.record_selector` / field selectors no longer match
2. **SPA timing** — headless fetch returns before JS renders the permit list
3. **Pagination change** — `pagination` settings out of sync with the site
4. **HTTP 403/429** — blocking or rate limits
5. **API/backend change** — HTML vs JSON or endpoint moved
6. **Silent zero rows** — scrape succeeds but finds no records

## Zero-record repair policy

Repair via `internal/repair` is **not** triggered on the first scrape that returns zero rows alone. Counties are marked `broken` and repaired when **`FindZeroCounties`** detects **two consecutive** successful scrapes with **zero** records (`internal/monitor/zerocheck.go`, daily cron in `internal/runner/cron.go`). That aligns production behavior with the existing monitor and avoids firing Claude Code on a single flaky empty run.

## How to verify changes

Apply migrations locally:

```bash
go run ./cmd/migrate
```

### Single county scrape

```bash
go run ./cmd/scrape-one --county=<county_id>
go run ./cmd/scrape-one --county=<county_id> -verbose
```

Expected: `found=N inserted=N duration=…`. A connector fix looks good when `found > 0`.

### Live page diagnostics

```bash
go run ./cmd/diagnose --county=<county_id>
```

### Tests and lint

```bash
go test ./...
go vet ./...
```

## Fix rules

1. **Connector config only** (DB updates to `connector_config`, status): preferably document in git via empty commit message if policy requires audit; production DB updates are often applied out-of-band — follow team process.
2. **Go changes**: open a PR; do not silently rewrite wide areas of unrelated code.
3. Run **`go run ./cmd/scrape-one`** where relevant before merging connector-facing work.
4. Run **`go test ./...`** and **`go vet ./...`** before merging Go changes.
5. **Never rewrite migrations** that already ran in production; add new numbered files under `migrations/`.

## Claude Code / Railway Docker

Production installs `@anthropic-ai/claude-code` (npm global) and copies `CLAUDE.md` plus `.claude/` into `/app`. **`GLASSBASE_CLI_WRAPPER=binary`** makes repair prompts invoke **`./scrape-one`**, **`./diagnose`**, **`./onboard`** (built binaries beside `glassbase`). Unset locally to keep using **`go run ./cmd/...`**.

`ANTHROPIC_API_KEY` is required. Optionally set **`CLAUDE_BIN`** or **`CLAUDE_WORKDIR`**.

Git is **not** assumed in the container: prioritize DB connector fixes and PRs opened from CI or a developer machine—not `git push` from Railway unless you wire credentials deliberately.

## Onboarding a county from scratch

```bash
go run ./cmd/onboard --url="<permit portal url>" --county="<slug>" --name="<display name>" --state=<ST>
```

## Commit messages

Example:

```
fix(county): broward_fl record_selector updated — table markup changed

- Selector: ...
- Verified: scrape-one found=N
```

## Environment

- Go 1.25+ (see `go.mod`)
- PostgreSQL (see existing migrations)
- Node 20 (frontend build; Claude Code in Docker image)
