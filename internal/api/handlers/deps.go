package handlers

import (
	"github.com/echayko/glassbase_permit_data_provider/config"
	"github.com/echayko/glassbase_permit_data_provider/internal/runner"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Deps bundles dependencies for HTTP handlers.
type Deps struct {
	Pool      *pgxpool.Pool
	Scheduler *runner.Scheduler
	Config    *config.Config
}
