package handlers

import (
	"context"
	"fmt"
	"regexp"

	"github.com/echayko/glassbase_permit_data_provider/internal/scraper"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var safePermitsTableRe = regexp.MustCompile(`^permits_[a-z0-9_]+$`)

func permitsTableIdent(countyID string) string {
	return scraper.PermitTableName(countyID)
}

func permitsTableExists(ctx context.Context, pool *pgxpool.Pool, countyID string) (bool, error) {
	tbl := permitsTableIdent(countyID)
	if !safePermitsTableRe.MatchString(tbl) {
		return false, nil
	}
	var exists bool
	err := pool.QueryRow(ctx,
		`SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = $1
		)`, tbl).Scan(&exists)
	return exists, err
}

func permitsCount(ctx context.Context, pool *pgxpool.Pool, countyID string) (int64, error) {
	exists, err := permitsTableExists(ctx, pool, countyID)
	if err != nil || !exists {
		return 0, err
	}
	tbl := permitsTableIdent(countyID)
	ident := pgx.Identifier{tbl}.Sanitize()
	var n int64
	q := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, ident)
	err = pool.QueryRow(ctx, q).Scan(&n)
	return n, err
}
