package registry

import (
	"encoding/json"
	"time"
)

type CountyConnector struct {
	ID              int             `db:"id" json:"id"`
	CountyID        string          `db:"county_id" json:"county_id"`
	CountyName      string          `db:"county_name" json:"county_name"`
	State           string          `db:"state" json:"state"`
	URL             string          `db:"url" json:"url"`
	SourceType      string          `db:"source_type" json:"source_type"`
	ConnectorConfig json.RawMessage `db:"connector_config" json:"connector_config"`
	Status          string          `db:"status" json:"status"`
	LastGeneratedAt *time.Time      `db:"last_generated_at" json:"last_generated_at"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at" json:"updated_at"`
}
