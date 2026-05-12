package monitor

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ZeroResult struct {
	CountyID         string
	LastRunStatus    string
	RecordsFound     int
	ConsecutiveZeros int
}

// FindZeroCounties returns counties where the last N finished runs were all successful with 0 records.
func FindZeroCounties(ctx context.Context, pool *pgxpool.Pool, consecutiveThreshold int) ([]ZeroResult, error) {
	if consecutiveThreshold < 1 {
		consecutiveThreshold = 2
	}

	const q = `
WITH ranked AS (
	SELECT county_id, status, records_found,
	       ROW_NUMBER() OVER (PARTITION BY county_id ORDER BY started_at DESC) AS rn
	FROM scrape_runs
	WHERE finished_at IS NOT NULL
),
lastn AS (
	SELECT county_id, status, records_found, rn
	FROM ranked
	WHERE rn <= $1
)
SELECT county_id,
       MAX(status) FILTER (WHERE rn = 1),
       COALESCE(MAX(records_found) FILTER (WHERE rn = 1), 0)::int,
       $1::int
FROM lastn
GROUP BY county_id
HAVING COUNT(*) = $1
   AND BOOL_AND(status = 'success')
   AND SUM(records_found) = 0`

	rows, err := pool.Query(ctx, q, consecutiveThreshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ZeroResult
	for rows.Next() {
		var r ZeroResult
		if err := rows.Scan(&r.CountyID, &r.LastRunStatus, &r.RecordsFound, &r.ConsecutiveZeros); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
