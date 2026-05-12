package repair

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/echayko/glassbase_permit_data_provider/internal/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RepairStreamLine is one line emitted by the Claude Code CLI while a repair runs.
// Progress and tool status usually appear on stderr; structured / print output on stdout.
type RepairStreamLine struct {
	Stream string // "stdout" or "stderr"
	Line   string
}

// Runner shells out to the Claude Code CLI.
type Runner struct {
	pool    *pgxpool.Pool
	workDir string
	timeout time.Duration
}

// NewRunner configures a repair runner with sane defaults.
func NewRunner(pool *pgxpool.Pool) *Runner {
	wd := os.Getenv("CLAUDE_WORKDIR")
	if wd == "" {
		if ex, err := os.Executable(); err == nil {
			wd = filepath.Dir(ex)
			if filepath.Base(wd) == "bin" {
				if p, err := filepath.Abs(filepath.Join(wd, "..")); err == nil {
					wd = p
				}
			}
		}
	}
	return &Runner{
		pool:    pool,
		workDir: wd,
		timeout: 20 * time.Minute,
	}
}

// RepairOutcome is the result of a single autonomous repair run.
type RepairOutcome struct {
	Success   bool
	Output    string
	CommitSHA string
	PRUrl     string
	Duration  time.Duration
	RunID     int
}

// RepairCounty runs Claude Code with a built-in repair prompt.
func (r *Runner) RepairCounty(ctx context.Context, county *registry.CountyConnector, trigger string) (*RepairOutcome, error) {
	return r.runClaudeRepair(ctx, county, trigger, nil)
}

// RepairCountyStreaming runs Claude Code and emits each stdout/stderr line via onLine before completion.
func (r *Runner) RepairCountyStreaming(ctx context.Context, county *registry.CountyConnector, trigger string, onLine func(RepairStreamLine)) (*RepairOutcome, error) {
	return r.runClaudeRepair(ctx, county, trigger, onLine)
}

func (r *Runner) runClaudeRepair(ctx context.Context, county *registry.CountyConnector, trigger string, onLine func(RepairStreamLine)) (*RepairOutcome, error) {
	if county == nil {
		return nil, fmt.Errorf("nil county")
	}

	lim := LimitsFromEnv()

	runsLog := r.loadRecentRuns(ctx, county.CountyID, lim)
	lastErr := truncateUTF8Runes(strings.TrimSpace(r.loadLastError(ctx, county.CountyID)), lim.MaxLastErrRunes)
	prompt := BuildPrompt(RepairContext{
		County:    county,
		Trigger:   trigger,
		LastError: lastErr,
		RunsLog:   runsLog,
	}, lim)

	ctxRun, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	cmd, cmdErr := claudeRepairCommand(ctxRun, prompt)
	if cmdErr != nil {
		return nil, cmdErr
	}

	runID, err := AcquireLock(ctx, r.pool, county.CountyID, trigger)
	if err != nil {
		log.Printf("[repair-ai] lock not acquired county=%s trigger=%s: %v", county.CountyID, trigger, err)
		return nil, fmt.Errorf("lock: %w", err)
	}

	outcome := &RepairOutcome{RunID: runID}
	workdirMsg := "(unset)"
	if r.workDir != "" {
		workdirMsg = r.workDir
	}
	log.Printf("[repair-ai] starting county=%s trigger=%s run_id=%d workdir=%s timeout=%v",
		county.CountyID, trigger, runID, workdirMsg, r.timeout)
	if r.workDir != "" {
		cmd.Dir = r.workDir
	}

	start := time.Now()

	var stderrBuf bytes.Buffer
	var combined strings.Builder

	if onLine == nil {
		var stdoutBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
		runErr := cmd.Run()
		outcome.Duration = time.Since(start)
		combined.WriteString(stdoutBuf.String())
		appendStderr(&combined, stderrBuf.Bytes())
		return r.finishRepairRun(ctx, county.CountyID, runID, outcome, combined.String(), runErr)
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("[repair-ai] county=%s run_id=%d pre-start failed (stdout pipe): %v", county.CountyID, runID, err)
		_ = FinishLock(ctx, r.pool, runID, "failed", "", "", "", fmt.Sprintf("stdout pipe: %v", err))
		return outcome, fmt.Errorf("stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		_ = stdoutPipe.Close()
		log.Printf("[repair-ai] county=%s run_id=%d pre-start failed (stderr pipe): %v", county.CountyID, runID, err)
		_ = FinishLock(ctx, r.pool, runID, "failed", "", "", "", fmt.Sprintf("stderr pipe: %v", err))
		return outcome, fmt.Errorf("stderr pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		_ = stdoutPipe.Close()
		_ = stderrPipe.Close()
		log.Printf("[repair-ai] county=%s run_id=%d pre-start failed (start): %v", county.CountyID, runID, err)
		_ = FinishLock(ctx, r.pool, runID, "failed", "", "", "", err.Error())
		return outcome, fmt.Errorf("start: %w", err)
	}

	const maxScan = 512 * 1024

	stderrDone := make(chan error, 1)
	go func() {
		scanErr := drainPipeLines(stderrPipe, &stderrBuf, maxScan, func(line string) {
			onLine(RepairStreamLine{Stream: "stderr", Line: line})
		})
		stderrDone <- scanErr
	}()

	stdoutScanErr := drainPipeLines(stdoutPipe, &combined, maxScan, func(line string) {
		onLine(RepairStreamLine{Stream: "stdout", Line: line})
	})
	stderrScanErr := <-stderrDone
	waitErr := cmd.Wait()

	runErr := waitErr
	if runErr == nil && stdoutScanErr != nil {
		runErr = stdoutScanErr
	}
	if runErr == nil && stderrScanErr != nil {
		runErr = stderrScanErr
	}
	outcome.Duration = time.Since(start)

	appendStderr(&combined, stderrBuf.Bytes())

	return r.finishRepairRun(ctx, county.CountyID, runID, outcome, combined.String(), runErr)
}

// drainPipeLines reads newline-delimited text from r, writes each line and '\n' to w, and calls emit per line.
func drainPipeLines(r io.Reader, w io.Writer, maxScan int, emit func(line string)) error {
	scan := bufio.NewScanner(r)
	buf := make([]byte, maxScan)
	scan.Buffer(buf, maxScan)
	for scan.Scan() {
		line := scan.Text()
		if _, err := io.WriteString(w, line); err != nil {
			return err
		}
		if _, err := w.Write([]byte{'\n'}); err != nil {
			return err
		}
		emit(line)
	}
	return scan.Err()
}

func appendStderr(b *strings.Builder, stderr []byte) {
	se := strings.TrimSpace(string(stderr))
	if se == "" {
		return
	}
	if b.Len() > 0 {
		b.WriteString("\n--- stderr ---\n")
	}
	b.WriteString(se)
}

func coalesceDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func logErrSnippet(msg string) string {
	if msg == "" {
		return "-"
	}
	const max = 400
	runes := []rune(msg)
	if len(runes) <= max {
		return msg
	}
	return string(runes[:max]) + "…"
}

func (r *Runner) finishRepairRun(ctx context.Context, countyID string, runID int, outcome *RepairOutcome, combinedOut string, runErr error) (*RepairOutcome, error) {
	outcome.Output = combinedOut
	outcome.Success = ParseSuccessFromOutput(combinedOut, runErr)
	outcome.CommitSHA = ExtractCommitSHA(combinedOut)
	outcome.PRUrl = ExtractPRURL(combinedOut)

	status := "success"
	errMsg := ""
	if runErr != nil {
		status = "failed"
		errMsg = runErr.Error()
		log.Printf("[repair-ai] Claude Code process exited with error county=%s run_id=%d: %v", countyID, runID, runErr)
	} else if !outcome.Success {
		status = "failed"
		errMsg = "repair heuristics did not detect success"
		log.Printf("[repair-ai] county=%s run_id=%d process exited 0 but output did not match success heuristics (see REPAIR SUMMARY / paused markers)", countyID, runID)
	}

	if finErr := FinishLock(ctx, r.pool, runID, status,
		combinedOut, outcome.CommitSHA, outcome.PRUrl, errMsg); finErr != nil {
		log.Printf("[repair-ai] FinishLock failed run_id=%d: %v", runID, finErr)
	}
	log.Printf("[repair-ai] finished county=%s run_id=%d status=%s duration=%v success=%v commit_sha=%s pr_url=%s err=%s",
		countyID, runID, status, outcome.Duration, outcome.Success,
		coalesceDash(outcome.CommitSHA), coalesceDash(outcome.PRUrl), logErrSnippet(errMsg))

	combinedLower := strings.ToLower(combinedOut)
	errLower := strings.ToLower(errMsg)
	if strings.Contains(combinedLower, "rate limit") ||
		strings.Contains(combinedLower, "tokens per minute") ||
		strings.Contains(errLower, "rate limit") {
		log.Printf("[repair-ai] rate limit (429/TPM): reduce repair prompt (GLASSBASE_REPAIR_*_RUNES), set GLASSBASE_REPAIR_COMPACT_PROMPT=true, or space cron runs (GLASSBASE_REPAIR_CRON_STAGGER_SECONDS=75); retry after ~1m. Docs: https://docs.claude.com/en/api/rate-limits")
	}

	if runErr != nil {
		return outcome, fmt.Errorf("claude code: %w", runErr)
	}
	return outcome, nil
}

func (r *Runner) loadRecentRuns(ctx context.Context, countyID string, lim EnvLimits) string {
	rows, err := r.pool.Query(ctx, `
		SELECT status, records_found, started_at, error_message
		FROM scrape_runs
		WHERE county_id = $1
		ORDER BY started_at DESC LIMIT 3`, countyID)
	if err != nil {
		return ""
	}
	defer rows.Close()

	var lines []string
	for rows.Next() {
		var status, errMsg string
		var found int
		var startedAt time.Time
		if err := rows.Scan(&status, &found, &startedAt, &errMsg); err != nil {
			continue
		}
		line := fmt.Sprintf("  %s: status=%s found=%d",
			startedAt.Format("2006-01-02 15:04"), status, found)
		if errMsg != "" {
			errMsg = truncateUTF8Runes(errMsg, lim.MaxRunErrRunes)
			line += " error=" + errMsg
		}
		lines = append(lines, line)
	}
	return truncateUTF8Runes(strings.Join(lines, "\n"), lim.MaxRunsLogRunes)
}

func (r *Runner) loadLastError(ctx context.Context, countyID string) string {
	var errMsg string
	_ = r.pool.QueryRow(ctx, `
		SELECT COALESCE(error_message, '')
		FROM scrape_runs
		WHERE county_id = $1 AND status = 'failed'
		ORDER BY started_at DESC LIMIT 1`, countyID).Scan(&errMsg)
	return errMsg
}
