package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/echayko/glassbase_permit_data_provider/internal/monitor"
	"github.com/echayko/glassbase_permit_data_provider/internal/onboard"
	"github.com/echayko/glassbase_permit_data_provider/internal/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

type countyPublic struct {
	CountyID          string     `json:"county_id"`
	CountyName        string     `json:"county_name"`
	State             string     `json:"state"`
	URL               string     `json:"url"`
	Status            string     `json:"status"`
	SourceType        string     `json:"source_type"`
	PermitCount       int64      `json:"permit_count"`
	LastRunAt         *time.Time `json:"last_run_at"`
	LastRunStatus     *string    `json:"last_run_status"`
	LastRunInserted   int        `json:"last_run_inserted"`
	LastGeneratedAt   *time.Time `json:"last_generated_at"`
}

type countiesListResponse struct {
	Counties []countyPublic `json:"counties"`
	Total    int            `json:"total"`
}

type addCountyBody struct {
	CountyID   string `json:"county_id"`
	CountyName string `json:"county_name"`
	State      string `json:"state"`
	URL        string `json:"url"`
}

type patchStatusBody struct {
	Status string `json:"status"`
}

type lastRunInfo struct {
	Status          string
	RecordsInserted int
	StartedAt       time.Time
}

func lastRunsByCounty(ctx context.Context, pool *pgxpool.Pool) (map[string]lastRunInfo, error) {
	out := map[string]lastRunInfo{}

	rows, err := pool.Query(ctx, `
		SELECT DISTINCT ON (county_id) county_id, status, records_inserted, started_at
		FROM scrape_runs
		ORDER BY county_id, started_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var countyID string
		var lr lastRunInfo
		if err := rows.Scan(&countyID, &lr.Status, &lr.RecordsInserted, &lr.StartedAt); err != nil {
			return nil, err
		}
		out[countyID] = lr
	}
	return out, rows.Err()
}

func countyToPublic(c registry.CountyConnector, permitCount int64, lr *lastRunInfo) countyPublic {
	p := countyPublic{
		CountyID:        c.CountyID,
		CountyName:      c.CountyName,
		State:           c.State,
		URL:             c.URL,
		Status:          c.Status,
		SourceType:      c.SourceType,
		PermitCount:     permitCount,
		LastRunInserted: 0,
		LastGeneratedAt: c.LastGeneratedAt,
	}
	if lr != nil {
		st := lr.Status
		p.LastRunStatus = &st
		t := lr.StartedAt
		p.LastRunAt = &t
		p.LastRunInserted = lr.RecordsInserted
	}
	return p
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}

// ListCounties returns visible counties with aggregated permit counts and last run info.
func (d *Deps) ListCounties() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		ctx := r.Context()
		store := registry.NewStore(d.Pool)
		all, err := store.GetAllVisible(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		lastMap, err := lastRunsByCounty(ctx, d.Pool)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		out := countiesListResponse{Counties: []countyPublic{}, Total: len(all)}
		for _, c := range all {
			pc, err := permitsCount(ctx, d.Pool, c.CountyID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			lr, ok := lastMap[c.CountyID]
			var lrPtr *lastRunInfo
			if ok {
				lrPtr = &lr
			}
			cp := countyToPublic(c, pc, lrPtr)
			out.Counties = append(out.Counties, cp)
		}

		writeJSON(w, http.StatusOK, out)
	}
}

// AddCounty runs AI onboarding for a new county connector.
func (d *Deps) AddCounty() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body addCountyBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if body.CountyID == "" || body.CountyName == "" || body.State == "" || body.URL == "" {
			http.Error(w, "county_id, county_name, state, url required", http.StatusBadRequest)
			return
		}

		if err := onboard.Run(r.Context(), d.Pool, d.Config, body.CountyID, body.CountyName, body.State, body.URL); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		writeJSON(w, http.StatusCreated, map[string]string{"county_id": body.CountyID})
	}
}

// GetCounty returns one county by ID.
func (d *Deps) GetCounty() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := r.PathValue("id")
		ctx := r.Context()
		store := registry.NewStore(d.Pool)
		c, err := store.GetByCountyID(ctx, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if c == nil || c.Status == "deleted" {
			http.NotFound(w, r)
			return
		}

		lastMap, err := lastRunsByCounty(ctx, d.Pool)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		pc, err := permitsCount(ctx, d.Pool, c.CountyID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		lr, ok := lastMap[c.CountyID]
		var lrPtr *lastRunInfo
		if ok {
			lrPtr = &lr
		}
		writeJSON(w, http.StatusOK, countyToPublic(*c, pc, lrPtr))
	}
}

// SetCountyStatus updates county status (active | paused).
func (d *Deps) SetCountyStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := r.PathValue("id")
		var body patchStatusBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if body.Status != "active" && body.Status != "paused" {
			http.Error(w, "status must be active or paused", http.StatusBadRequest)
			return
		}

		store := registry.NewStore(d.Pool)
		c, err := store.GetByCountyID(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if c == nil || c.Status == "deleted" {
			http.NotFound(w, r)
			return
		}

		if err := store.SetStatus(r.Context(), id, body.Status); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": body.Status})
	}
}

// DeleteCounty soft-deletes a county (status = deleted).
func (d *Deps) DeleteCounty() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := r.PathValue("id")
		store := registry.NewStore(d.Pool)
		c, err := store.GetByCountyID(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if c == nil || c.Status == "deleted" {
			http.NotFound(w, r)
			return
		}
		if err := store.SetStatus(r.Context(), id, "deleted"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// TriggerCountyRun queues a manual scrape for one county.
func (d *Deps) TriggerCountyRun() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := r.PathValue("id")
		if err := d.Scheduler.RunCounty(id); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"status":"queued"}`))
	}
}

// RepairCounty regenerates connector config via Claude for an existing county.
func (d *Deps) RepairCounty() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := r.PathValue("id")
		cfg := d.Config
		if cfg == nil {
			http.Error(w, "config missing", http.StatusInternalServerError)
			return
		}
		rep := monitor.NewRepairer(d.Pool, cfg.AnthropicAPIKey, cfg.AnthropicModel)
		res := rep.RepairCounty(r.Context(), id)
		if !res.Success {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"reason":  res.Reason,
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success":   true,
			"county_id": id,
		})
	}
}
