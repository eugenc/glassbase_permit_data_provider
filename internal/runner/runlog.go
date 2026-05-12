package runner

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RunLog struct {
	pool *pgxpool.Pool
}

func NewRunLog(pool *pgxpool.Pool) *RunLog {
	return &RunLog{pool: pool}
}

// Start inserts a "running" row and returns its ID.
func (r *RunLog) Start(ctx context.Context, countyID string) (int, error) {
	var id int
	err := r.pool.QueryRow(ctx, `
		INSERT INTO scrape_runs (county_id, status, started_at)
		VALUES ($1, 'running', NOW())
		RETURNING id`, countyID).Scan(&id)
	return id, err
}

// Finish updates the run row with final status and stats.
func (r *RunLog) Finish(ctx context.Context, id int, status string, found, inserted int, errMsg string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE scrape_runs SET
			status           = $1,
			finished_at      = NOW(),
			records_found    = $2,
			records_inserted = $3,
			error_message    = NULLIF($4, '')
		WHERE id = $5`,
		status, found, inserted, errMsg, id)
	return err
}
