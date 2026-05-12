package api

import (
	"net/http"

	"github.com/echayko/glassbase_permit_data_provider/config"
	"github.com/echayko/glassbase_permit_data_provider/internal/api/handlers"
	"github.com/echayko/glassbase_permit_data_provider/internal/runner"
	"github.com/jackc/pgx/v5/pgxpool"
)

func corsMiddleware(origins []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		for _, o := range origins {
			if o == origin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// NewServer configures the admin HTTP API and embedded SPA.
func NewServer(pool *pgxpool.Pool, sched *runner.Scheduler, cfg *config.Config) *http.Server {
	mux := http.NewServeMux()
	jwtSecret := []byte(cfg.JWTSecret)

	mux.HandleFunc("POST /auth/login", HandleLogin(pool, jwtSecret))

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := pool.Ping(r.Context()); err != nil {
			http.Error(w, "db down", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	deps := &handlers.Deps{Pool: pool, Scheduler: sched, Config: cfg}
	apiMux := http.NewServeMux()

	apiMux.HandleFunc("GET /dashboard", deps.Dashboard())

	apiMux.HandleFunc("GET /repairs/recent", deps.RecentRepairs())

	apiMux.HandleFunc("GET /counties", deps.ListCounties())
	apiMux.HandleFunc("POST /counties", deps.AddCounty())

	apiMux.HandleFunc("GET /counties/{id}/permits/export", deps.ExportPermits())
	apiMux.HandleFunc("GET /counties/{id}/permits", deps.ListPermits())

	apiMux.HandleFunc("PATCH /counties/{id}/status", deps.SetCountyStatus())
	apiMux.HandleFunc("DELETE /counties/{id}", deps.DeleteCounty())
	apiMux.HandleFunc("POST /counties/{id}/run", deps.TriggerCountyRun())
	apiMux.HandleFunc("POST /counties/{id}/repair", deps.RepairCounty())
	apiMux.HandleFunc("POST /counties/{id}/repair-cc", deps.RepairWithClaudeCode())
	apiMux.HandleFunc("GET /counties/{id}", deps.GetCounty())

	apiMux.HandleFunc("GET /runs/error-rate", deps.ErrorRate())
	apiMux.HandleFunc("GET /runs", deps.ListRuns())

	mux.Handle("/api/", AuthMiddleware(jwtSecret, http.StripPrefix("/api", apiMux)))

	mux.Handle("/", StaticHandler())

	handler := corsMiddleware(cfg.CORSAllowedOrigins, mux)

	return &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: handler,
	}
}
