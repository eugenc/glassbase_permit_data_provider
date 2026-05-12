package onboard

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/echayko/glassbase_permit_data_provider/config"
	"github.com/echayko/glassbase_permit_data_provider/internal/fetcher"
	"github.com/echayko/glassbase_permit_data_provider/internal/generator"
	"github.com/echayko/glassbase_permit_data_provider/internal/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Run detects source type, generates connector config via Claude, and upserts the county row.
func Run(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config, countyID, countyName, state, url string) error {
	if countyID == "" || countyName == "" || state == "" || url == "" {
		return fmt.Errorf("county_id, county_name, state, and url are required")
	}

	sourceType, err := fetcher.DetectSourceType(ctx, url)
	if err != nil {
		return fmt.Errorf("detect source: %w", err)
	}

	f := fetcher.New(sourceType)
	result, err := f.Fetch(ctx, url)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	var networkSummary []string
	for _, nc := range result.NetworkCalls {
		m := 200
		if len(nc.Response) < m {
			m = len(nc.Response)
		}
		suffix := ""
		if m > 0 {
			suffix = nc.Response[:m]
		}
		networkSummary = append(networkSummary, nc.URL+" → "+suffix)
	}

	prompt := generator.BuildPrompt(url, sourceType, result.Body, networkSummary)
	gen := &generator.Generator{APIKey: cfg.AnthropicAPIKey, Model: cfg.AnthropicModel}
	connCfg, err := gen.GenerateConnectorConfig(ctx, prompt)
	if err != nil {
		return fmt.Errorf("generator: %w", err)
	}

	if connCfg.SourceType != "" {
		sourceType = connCfg.SourceType
	}

	configJSON, err := json.Marshal(connCfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	now := time.Now()
	store := registry.NewStore(pool)
	return store.Upsert(ctx, &registry.CountyConnector{
		CountyID:        countyID,
		CountyName:      countyName,
		State:           state,
		URL:             url,
		SourceType:      sourceType,
		ConnectorConfig: configJSON,
		Status:          "active",
		LastGeneratedAt: &now,
	})
}
