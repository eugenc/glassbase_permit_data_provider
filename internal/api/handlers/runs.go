package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type scrapeRunPublic struct {
	ID              int        `json:"id"`
	CountyID        string     `json:"county_id"`
	Status          string     `json:"status"`
	RecordsFound    int        `json:"records_found"`
	RecordsInserted int        `json:"records_inserted"`
	StartedAt       time.Time  `json:"started_at"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
	ErrorMessage    *string    `json:"error_message,omitempty"`
}

type runsListResponse struct {
	Runs []scrapeRunPublic `json:"runs"`
}

type errorRateRow struct {
	Day      string  `json:"day"`
	ErrorPct float64 `json:"error_pct"`
	Total    int64   `json:"total"`
	Failed   int64   `json:"failed"`
}

// ListRuns returns scrape history with optional filters.
func (d *Deps) ListRuns() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		q := r.URL.Query()
		countyID := q.Get("county_id")
		status := q.Get("status")
		limit := parseIntDefault(q.Get("limit"), 50)
		if limit > 200 {
			limit = 200
		}

		sqlQuery := `
			SELECT id, county_id, status, records_found, records_inserted,
			       started_at, finished_at, error_message
			FROM scrape_runs
			WHERE TRUE`
		args := []interface{}{}

		if countyID != "" {
			sqlQuery += fmt.Sprintf(` AND county_id = $%d`, len(args)+1)
			args = append(args, countyID)
		}
		if status != "" {
			sqlQuery += fmt.Sprintf(` AND status = $%d`, len(args)+1)
			args = append(args, status)
		}

		sqlQuery += fmt.Sprintf(`
			ORDER BY started_at DESC
			LIMIT $%d`, len(args)+1)
		args = append(args, limit)

		rows, err := d.Pool.Query(r.Context(), sqlQuery, args...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var runs []scrapeRunPublic
		for rows.Next() {
			var run scrapeRunPublic
			var finishedAt sql.NullTime
			var errMsg sql.NullString
			if err := rows.Scan(&run.ID, &run.CountyID, &run.Status,
				&run.RecordsFound, &run.RecordsInserted,
				&run.StartedAt, &finishedAt, &errMsg); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if finishedAt.Valid {
				t := finishedAt.Time
				run.FinishedAt = &t
			}
			if errMsg.Valid {
				s := errMsg.String
				run.ErrorMessage = &s
			}
			runs = append(runs, run)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, runsListResponse{Runs: runs})
	}
}

// ErrorRate returns daily failure percentages for the last 30 days.
func (d *Deps) ErrorRate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rows, err := d.Pool.Query(r.Context(), `
			SELECT
				DATE(started_at)::text AS day,
				COUNT(*) AS total,
				COUNT(*) FILTER (WHERE status = 'failed') AS failed,
				CASE WHEN COUNT(*) = 0 THEN 0
				     ELSE ROUND(100.0 * COUNT(*) FILTER (WHERE status = 'failed') / COUNT(*), 1)
				END AS error_pct
			FROM scrape_runs
			WHERE started_at >= NOW() - INTERVAL '30 days'
			GROUP BY DATE(started_at)
			ORDER BY day ASC`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var out []errorRateRow
		for rows.Next() {
			var row errorRateRow
			var pct float64
			if err := rows.Scan(&row.Day, &row.Total, &row.Failed, &pct); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			row.ErrorPct = pct
			out = append(out, row)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		_ = enc.Encode(out)
	}
}
