package repair

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AcquireLock inserts a running repair_runs row if no recent running repair exists for the county.
func AcquireLock(ctx context.Context, pool *pgxpool.Pool, countyID, trigger string) (int, error) {
	var existing int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM repair_runs
		WHERE county_id = $1 AND status = 'running'
		AND started_at > NOW() - INTERVAL '30 minutes'`,
		countyID).Scan(&existing)
	if err != nil {
		return 0, err
	}
	if existing > 0 {
		return 0, fmt.Errorf("repair already in progress for %s", countyID)
	}

	var id int
	err = pool.QueryRow(ctx, `
		INSERT INTO repair_runs (county_id, repair_trigger, status)
		VALUES ($1, $2, 'running')
		RETURNING id`,
		countyID, trigger).Scan(&id)
	return id, err
}

// FinishLock updates the repair_run row with final outcome.
func FinishLock(ctx context.Context, pool *pgxpool.Pool, id int,
	status, output, commitSHA, prURL, errMsg string) error {
	_, err := pool.Exec(ctx, `
		UPDATE repair_runs SET
			status        = $1,
			claude_output = $2,
			commit_sha    = $3,
			pr_url        = $4,
			error_message = NULLIF($5, ''),
			finished_at   = NOW()
		WHERE id = $6`,
		status, output, nullIfEmpty(commitSHA), nullIfEmpty(prURL), errMsg, id)
	return err
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
