package scraper

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// fixedPermitColumns are declared in CREATE TABLE; extraction field names must not collide.
var fixedPermitColumns = map[string]struct{}{
	"id":            {},
	"permit_number": {},
	"scraped_at":    {},
	"raw_data":      {},
}

func isFixedPermitColumn(name string) bool {
	_, ok := fixedPermitColumns[name]
	return ok
}

// EnsureTable creates permits_{countyID} (base columns only) if missing, then adds any connector
// field columns that are not yet present (handles connector regeneration / schema drift).
// Column names come from each field's connector `name` after sanitizeColName (dots and other
// punctuation become underscores, so names derived from JSON paths stay valid identifiers).
func EnsureTable(ctx context.Context, pool *pgxpool.Pool, countyID string, fields []string) error {
	tableName := tableNameFor(countyID)
	ukConstraint := "uk_" + sanitizeColName(countyID) + "_permit"

	createSQL := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
    id             SERIAL PRIMARY KEY,
    permit_number  TEXT NOT NULL,
    scraped_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    raw_data       JSONB NOT NULL DEFAULT '{}',
    CONSTRAINT %s UNIQUE (permit_number)
);`, tableName, ukConstraint)

	if _, err := pool.Exec(ctx, createSQL); err != nil {
		return fmt.Errorf("create table: %w", err)
	}

	existing, err := loadExistingColumns(ctx, pool, tableName)
	if err != nil {
		return err
	}

	seen := make(map[string]struct{})
	for _, f := range fields {
		col := sanitizeColName(f)
		if isFixedPermitColumn(col) {
			continue
		}
		if _, dup := seen[col]; dup {
			continue
		}
		seen[col] = struct{}{}
		if _, ok := existing[col]; ok {
			continue
		}
		alter := fmt.Sprintf(`ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s TEXT`, tableName, col)
		if _, err := pool.Exec(ctx, alter); err != nil {
			return fmt.Errorf("add column %s: %w", col, err)
		}
		log.Printf("[schema] added column %s to %s", col, tableName)
	}

	ffSQL := fmt.Sprintf(`ALTER TABLE %s SET (fillfactor = 90)`, tableName)
	if _, err := pool.Exec(ctx, ffSQL); err != nil {
		return err
	}

	idxName := "idx_" + sanitizeColName(countyID) + "_scraped_at"
	idxSQL := fmt.Sprintf(
		`CREATE INDEX IF NOT EXISTS %s ON %s(scraped_at DESC);`,
		idxName, tableName,
	)
	if _, err := pool.Exec(ctx, idxSQL); err != nil {
		return err
	}
	return nil
}

func loadExistingColumns(ctx context.Context, pool *pgxpool.Pool, tableName string) (map[string]struct{}, error) {
	rows, err := pool.Query(ctx, `
		SELECT column_name
		FROM information_schema.columns
		WHERE table_schema = current_schema()
		  AND table_name = $1`,
		tableName)
	if err != nil {
		return nil, fmt.Errorf("list columns for %s: %w", tableName, err)
	}
	defer rows.Close()

	out := make(map[string]struct{})
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			return nil, err
		}
		out[strings.ToLower(col)] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func tableNameFor(countyID string) string {
	return "permits_" + sanitizeColName(countyID)
}

// PermitTableName returns the PostgreSQL table name for a county's permits (sanitized).
func PermitTableName(countyID string) string {
	return tableNameFor(countyID)
}

func sanitizeColName(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	out := b.String()
	if out == "" {
		return "col"
	}
	return out
}
