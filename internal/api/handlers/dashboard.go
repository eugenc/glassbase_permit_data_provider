package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/echayko/glassbase_permit_data_provider/internal/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

type statusCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

type dayCount struct {
	Day   string `json:"day"`
	Count int64  `json:"count"`
}

type recentError struct {
	CountyID     string `json:"county_id"`
	ErrorMessage string `json:"error_message"`
	StartedAt    string `json:"started_at"`
}

type dashboardStats struct {
	TotalCounties   int           `json:"total_counties"`
	ActiveCounties  int           `json:"active_counties"`
	BrokenCounties  int           `json:"broken_counties"`
	PausedCounties  int           `json:"paused_counties"`
	PermitsThisWeek int64         `json:"permits_this_week"`
	TotalPermits    int64         `json:"total_permits"`
	LastRunAt       *time.Time    `json:"last_run_at"`
	RunsThisWeek    int64         `json:"runs_this_week"`
	ErrorRate       float64       `json:"error_rate"`
	ByStatus        []statusCount `json:"by_status"`
	PermitsByDay    []dayCount    `json:"permits_by_day"`
	RecentErrors    []recentError `json:"recent_errors"`
}

// Dashboard returns aggregated stats for the admin home page.
func (d *Deps) Dashboard() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		stats, err := buildDashboardStats(r.Context(), d.Pool)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		_ = enc.Encode(stats)
	}
}

func buildDashboardStats(ctx context.Context, pool *pgxpool.Pool) (*dashboardStats, error) {
	out := &dashboardStats{
		ByStatus:     []statusCount{},
		PermitsByDay: []dayCount{},
		RecentErrors: []recentError{},
	}

	store := registry.NewStore(pool)
	counties, err := store.GetAllVisible(ctx)
	if err != nil {
		return nil, err
	}

	statusTotals := map[string]int{}
	for _, c := range counties {
		statusTotals[c.Status]++
	}
	out.TotalCounties = len(counties)
	out.ActiveCounties = statusTotals["active"]
	out.BrokenCounties = statusTotals["broken"]
	out.PausedCounties = statusTotals["paused"]

	for st, n := range statusTotals {
		out.ByStatus = append(out.ByStatus, statusCount{Status: st, Count: n})
	}

	var totalPermits int64
	for _, c := range counties {
		n, err := permitsCount(ctx, pool, c.CountyID)
		if err != nil {
			return nil, err
		}
		totalPermits += n
	}
	out.TotalPermits = totalPermits

	err = pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(records_inserted), 0)
		FROM scrape_runs
		WHERE started_at >= NOW() - INTERVAL '7 days'
		  AND status = 'success'`).Scan(&out.PermitsThisWeek)
	if err != nil {
		return nil, err
	}

	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM scrape_runs
		WHERE started_at >= NOW() - INTERVAL '7 days'`).Scan(&out.RunsThisWeek)
	if err != nil {
		return nil, err
	}

	var failedWeek, totalWeek sql.NullInt64
	err = pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE status = 'failed'),
			COUNT(*)
		FROM scrape_runs
		WHERE started_at >= NOW() - INTERVAL '7 days'`).Scan(&failedWeek, &totalWeek)
	if err != nil {
		return nil, err
	}
	if totalWeek.Valid && totalWeek.Int64 > 0 {
		out.ErrorRate = float64(failedWeek.Int64) / float64(totalWeek.Int64) * 100
	}

	var lastRun sql.NullTime
	err = pool.QueryRow(ctx, `SELECT MAX(started_at) FROM scrape_runs`).Scan(&lastRun)
	if err != nil {
		return nil, err
	}
	if lastRun.Valid {
		t := lastRun.Time
		out.LastRunAt = &t
	}

	rows, err := pool.Query(ctx, `
		SELECT DATE(started_at)::text AS day, COALESCE(SUM(records_inserted), 0)::bigint AS cnt
		FROM scrape_runs
		WHERE status = 'success'
		  AND started_at >= NOW() - INTERVAL '7 days'
		GROUP BY DATE(started_at)
		ORDER BY day ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	perDay := map[string]int64{}
	for rows.Next() {
		var day string
		var cnt int64
		if err := rows.Scan(&day, &cnt); err != nil {
			return nil, err
		}
		perDay[day] = cnt
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	for i := 6; i >= 0; i-- {
		d := now.Truncate(24*time.Hour).AddDate(0, 0, -i).Format("2006-01-02")
		out.PermitsByDay = append(out.PermitsByDay, dayCount{Day: d, Count: perDay[d]})
	}

	errRows, err := pool.Query(ctx, `
		SELECT county_id, COALESCE(error_message, ''), started_at
		FROM scrape_runs
		WHERE status = 'failed'
		  AND started_at >= NOW() - INTERVAL '7 days'
		ORDER BY started_at DESC
		LIMIT 50`)
	if err != nil {
		return nil, err
	}
	defer errRows.Close()

	for errRows.Next() {
		var countyID, msg string
		var started time.Time
		if err := errRows.Scan(&countyID, &msg, &started); err != nil {
			return nil, err
		}
		out.RecentErrors = append(out.RecentErrors, recentError{
			CountyID:     countyID,
			ErrorMessage: msg,
			StartedAt:    started.UTC().Format(time.RFC3339),
		})
	}

	return out, errRows.Err()
}
