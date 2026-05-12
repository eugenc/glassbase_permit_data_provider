package registry

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func scanCounty(row pgx.Row) (CountyConnector, error) {
	var c CountyConnector
	err := row.Scan(
		&c.ID,
		&c.CountyID,
		&c.CountyName,
		&c.State,
		&c.URL,
		&c.SourceType,
		&c.ConnectorConfig,
		&c.Status,
		&c.LastGeneratedAt,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	return c, err
}

func (s *Store) GetAll(ctx context.Context) ([]CountyConnector, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, county_id, county_name, state, url, source_type,
			connector_config, status, last_generated_at, created_at, updated_at
		 FROM county_connectors ORDER BY state, county_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []CountyConnector
	for rows.Next() {
		c, err := scanCounty(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) GetByCountyID(ctx context.Context, countyID string) (*CountyConnector, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, county_id, county_name, state, url, source_type,
			connector_config, status, last_generated_at, created_at, updated_at
		 FROM county_connectors WHERE county_id = $1`, countyID)
	c, err := scanCounty(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (s *Store) GetByStatus(ctx context.Context, status string) ([]CountyConnector, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, county_id, county_name, state, url, source_type,
			connector_config, status, last_generated_at, created_at, updated_at
		 FROM county_connectors WHERE status = $1 ORDER BY state, county_name`,
		status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []CountyConnector
	for rows.Next() {
		c, err := scanCounty(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) Upsert(ctx context.Context, c *CountyConnector) error {
	if c == nil {
		return fmt.Errorf("nil county connector")
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO county_connectors
			(county_id, county_name, state, url, source_type, connector_config, status, last_generated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (county_id) DO UPDATE SET
			county_name       = EXCLUDED.county_name,
			state             = EXCLUDED.state,
			url               = EXCLUDED.url,
			connector_config  = EXCLUDED.connector_config,
			source_type       = EXCLUDED.source_type,
			status            = EXCLUDED.status,
			last_generated_at = EXCLUDED.last_generated_at,
			updated_at        = NOW()`,
		c.CountyID, c.CountyName, c.State, c.URL,
		c.SourceType, c.ConnectorConfig, c.Status, c.LastGeneratedAt,
	)
	return err
}

func (s *Store) SetStatus(ctx context.Context, countyID, status string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE county_connectors SET status=$1, updated_at=NOW() WHERE county_id=$2`,
		status, countyID)
	return err
}
