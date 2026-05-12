package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/echayko/glassbase_permit_data_provider/internal/registry"
	"github.com/echayko/glassbase_permit_data_provider/internal/repair"
)

type repairRow struct {
	ID            int        `json:"id"`
	CountyID      string     `json:"county_id"`
	Trigger       string     `json:"trigger"`
	Status        string     `json:"status"`
	CommitSHA     *string    `json:"commit_sha"`
	PRUrl         *string    `json:"pr_url"`
	StartedAt     time.Time  `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at"`
	ErrorMessage  *string    `json:"error_message"`
	OutputPreview string     `json:"output_preview"`
}

// repairDetail is a single repair run with full Claude Code transcript (GET /repairs/{id}/log).
type repairDetail struct {
	ID            int        `json:"id"`
	CountyID      string     `json:"county_id"`
	Trigger       string     `json:"trigger"`
	Status        string     `json:"status"`
	CommitSHA     *string    `json:"commit_sha"`
	PRUrl         *string    `json:"pr_url"`
	StartedAt     time.Time  `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at"`
	ErrorMessage  *string    `json:"error_message"`
	ClaudeOutput  string     `json:"claude_output"`
}

// RecentRepairs returns the latest Claude Code repair attempts.
func (d *Deps) RecentRepairs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		rows, err := d.Pool.Query(r.Context(), `
			SELECT id, county_id, repair_trigger, status, commit_sha, pr_url,
			       started_at, finished_at, error_message,
			       LEFT(claude_output, 800)::text AS output_preview
			FROM repair_runs
			ORDER BY started_at DESC
			LIMIT 20`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var repairs []repairRow
		for rows.Next() {
			var rr repairRow
			if err := rows.Scan(&rr.ID, &rr.CountyID, &rr.Trigger, &rr.Status,
				&rr.CommitSHA, &rr.PRUrl, &rr.StartedAt, &rr.FinishedAt,
				&rr.ErrorMessage, &rr.OutputPreview); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			repairs = append(repairs, rr)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"repairs": repairs})
	}
}

// RepairRunLog returns one repair_run row including full claude_output.
func (d *Deps) RepairRunLog() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil || id < 1 {
			http.Error(w, "invalid repair id", http.StatusBadRequest)
			return
		}

		var rd repairDetail
		var commit, prURL, errMsg *string
		var finished *time.Time

		rowErr := d.Pool.QueryRow(r.Context(), `
			SELECT id, county_id, repair_trigger, status, commit_sha, pr_url,
			       started_at, finished_at, error_message,
			       COALESCE(claude_output, '')
			FROM repair_runs
			WHERE id = $1`, id).Scan(&rd.ID, &rd.CountyID, &rd.Trigger, &rd.Status,
			&commit, &prURL, &rd.StartedAt, &finished, &errMsg, &rd.ClaudeOutput)

		if rowErr != nil {
			if errors.Is(rowErr, pgx.ErrNoRows) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, rowErr.Error(), http.StatusInternalServerError)
			return
		}
		rd.CommitSHA = commit
		rd.PRUrl = prURL
		rd.FinishedAt = finished
		rd.ErrorMessage = errMsg

		writeJSON(w, http.StatusOK, rd)
	}
}

// RepairWithClaudeCode streams Claude Code stdout as SSE (POST so clients can attach Authorization headers).
func (d *Deps) RepairWithClaudeCode() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		countyID := r.PathValue("id")
		store := registry.NewStore(d.Pool)
		c, err := store.GetByCountyID(r.Context(), countyID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if c == nil {
			http.NotFound(w, r)
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		h := w.Header()
		h.Set("Content-Type", "text/event-stream")
		h.Set("Cache-Control", "no-cache")
		h.Set("Connection", "keep-alive")
		h.Set("X-Accel-Buffering", "no")

		sendJSON := func(event string, v interface{}) {
			b, _ := json.Marshal(v)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, string(b))
			flusher.Flush()
		}

		sendJSON("start", map[string]string{
			"county_id": countyID,
			"message":   "starting Claude Code repair",
		})

		sendJSON("pulse", map[string]interface{}{
			"elapsed_sec": 0,
			"message":     "Repair started — live lines appear after the CLI emits output (often stderr). Quiet stretches are normal.",
		})

		log.Printf("[repair-ai] API SSE repair started county=%s remote=%s", countyID, r.RemoteAddr)

		run := repair.NewRunner(d.Pool)

		pulseStop := make(chan struct{})
		var pulseWG sync.WaitGroup
		pulseWG.Add(1)
		go func() {
			defer pulseWG.Done()
			started := time.Now()
			ticker := time.NewTicker(3 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-pulseStop:
					return
				case <-ticker.C:
					sec := int(time.Since(started) / time.Second)
					sendJSON("pulse", map[string]interface{}{
						"elapsed_sec": sec,
						"message": fmt.Sprintf(
							"Claude Code still running (%ds) — streamed lines appear after the CLI prints a newline; "+
								"quiet UI during long turns is normal.",
							sec,
						),
					})
				}
			}
		}()

		outcome, execErr := run.RepairCountyStreaming(r.Context(), c, "manual", func(chunk repair.RepairStreamLine) {
			sendJSON("output", map[string]string{
				"line":   chunk.Line,
				"stream": chunk.Stream,
			})
		})

		close(pulseStop)
		pulseWG.Wait()
		storeAfter := registry.NewStore(d.Pool)

		switch {
		case outcome == nil && execErr != nil:
			errMsg := execErr.Error()
			if len(errMsg) > 400 {
				errMsg = errMsg[:400] + "…"
			}
			log.Printf("[repair-ai] API SSE repair ended county=%s outcome=nil err=%s", countyID, errMsg)
			sendJSON("error", map[string]string{"message": execErr.Error()})
		case outcome == nil:
			log.Printf("[repair-ai] API SSE repair ended county=%s outcome=nil err=nil (unexpected)", countyID)
			sendJSON("error", map[string]string{"message": "repair failed unexpectedly"})
		default:
			if outcome.Success {
				_ = storeAfter.SetStatus(r.Context(), countyID, "active")
				log.Printf("[repair-ai] API SSE repair complete county=%s success=true run_id=%d duration=%v commit_sha=%s pr_url=%s",
					countyID, outcome.RunID, outcome.Duration, outcome.CommitSHA, outcome.PRUrl)
				sendJSON("complete", map[string]interface{}{
					"success":    true,
					"commit_sha": outcome.CommitSHA,
					"pr_url":     outcome.PRUrl,
				})
				return
			}
			msg := "Claude Code could not repair automatically"
			if execErr != nil {
				msg = execErr.Error()
			}
			_ = storeAfter.SetStatus(r.Context(), countyID, "paused")
			logMsg := msg
			if len(logMsg) > 400 {
				logMsg = logMsg[:400] + "…"
			}
			log.Printf("[repair-ai] API SSE repair complete county=%s success=false run_id=%d duration=%v message=%s",
				countyID, outcome.RunID, outcome.Duration, logMsg)
			sendJSON("complete", map[string]interface{}{
				"success": false,
				"message": msg,
			})
		}
	}
}
