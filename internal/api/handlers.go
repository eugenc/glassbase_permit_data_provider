package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/echayko/glassbase_permit_data_provider/config"
	"github.com/echayko/glassbase_permit_data_provider/internal/monitor"
	"github.com/echayko/glassbase_permit_data_provider/internal/registry"
	"github.com/echayko/glassbase_permit_data_provider/internal/runner"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handlers struct {
	pool      *pgxpool.Pool
	scheduler *runner.Scheduler
	Config    *config.Config
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	if err := h.pool.Ping(r.Context()); err != nil {
		http.Error(w, "db down", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handlers) ListCounties(w http.ResponseWriter, r *http.Request) {
	store := registry.NewStore(h.pool)
	counties, err := store.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(counties)
}

func (h *Handlers) BrokenCounties(w http.ResponseWriter, r *http.Request) {
	store := registry.NewStore(h.pool)
	counties, err := store.GetByStatus(r.Context(), "broken")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(counties)
}

type runRow struct {
	CountyID        string     `json:"county_id"`
	Status          string     `json:"status"`
	RecordsFound    int        `json:"records_found"`
	RecordsInserted int        `json:"records_inserted"`
	StartedAt       time.Time  `json:"started_at"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
	ErrorMessage    *string    `json:"error_message,omitempty"`
}

func (h *Handlers) RecentRuns(w http.ResponseWriter, r *http.Request) {
	rows, err := h.pool.Query(r.Context(), `
		SELECT county_id, status, records_found, records_inserted,
		       started_at, finished_at, error_message
		FROM scrape_runs
		ORDER BY started_at DESC
		LIMIT 50`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var out []runRow
	for rows.Next() {
		var rr runRow
		var finishedAt sql.NullTime
		var errMsg sql.NullString
		if err := rows.Scan(&rr.CountyID, &rr.Status, &rr.RecordsFound, &rr.RecordsInserted,
			&rr.StartedAt, &finishedAt, &errMsg); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if finishedAt.Valid {
			t := finishedAt.Time
			rr.FinishedAt = &t
		}
		if errMsg.Valid {
			s := errMsg.String
			rr.ErrorMessage = &s
		}
		out = append(out, rr)
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

func (h *Handlers) TriggerRunNow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	go h.scheduler.RunNow()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"status":"started"}`))
}

func (h *Handlers) RepairCounty(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	cfg := h.Config
	if cfg == nil {
		http.Error(w, "config missing", http.StatusInternalServerError)
		return
	}
	rep := monitor.NewRepairer(h.pool, cfg.AnthropicAPIKey, cfg.AnthropicModel)
	res := rep.RepairCounty(r.Context(), id)
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if !res.Success {
		w.WriteHeader(http.StatusBadRequest)
		_ = enc.Encode(map[string]interface{}{
			"success": false,
			"reason":  res.Reason,
		})
		return
	}
	_ = enc.Encode(map[string]interface{}{
		"success":   true,
		"county_id": id,
	})
}
