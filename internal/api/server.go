package api

import (
	"net/http"

	"github.com/echayko/glassbase_permit_data_provider/config"
	"github.com/echayko/glassbase_permit_data_provider/internal/runner"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewServer configures the admin HTTP API.
func NewServer(pool *pgxpool.Pool, sched *runner.Scheduler, cfg *config.Config) *http.Server {
	mux := http.NewServeMux()
	h := &Handlers{pool: pool, scheduler: sched, Config: cfg}

	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /counties", h.ListCounties)
	mux.HandleFunc("GET /counties/broken", h.BrokenCounties)
	mux.HandleFunc("GET /runs/recent", h.RecentRuns)
	mux.HandleFunc("POST /run/now", h.TriggerRunNow)
	mux.HandleFunc("POST /repair/{id}", h.RepairCounty)

	return &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: mux,
	}
}
