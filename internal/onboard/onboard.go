package onboard

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

	log.Printf("onboard[%s] start url=%s", countyID, url)

	sourceType, err := fetcher.DetectSourceType(ctx, url)
	if err != nil {
		return fmt.Errorf("detect source: %w", err)
	}
	log.Printf("onboard[%s] detected source_type=%s", countyID, sourceType)

	f := fetcher.New(sourceType)
	result, err := f.Fetch(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	log.Printf("onboard[%s] fetch done: html_bytes=%d network_calls=%d", countyID, len(result.Body), len(result.NetworkCalls))

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
	log.Printf("onboard[%s] prompt built: runes=%d network_summary_lines=%d", countyID, len([]rune(prompt)), len(networkSummary))

	gen := &generator.Generator{APIKey: cfg.AnthropicAPIKey, Model: cfg.AnthropicModel}
	connCfg, err := gen.GenerateConnectorConfig(ctx, prompt)
	if err != nil {
		return fmt.Errorf("generator: %w", err)
	}
	log.Printf("onboard[%s] generator returned connector config (source_type in config=%q)", countyID, connCfg.SourceType)

	if connCfg.SourceType != "" {
		sourceType = connCfg.SourceType
	}

	configJSON, err := json.Marshal(connCfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	log.Printf("onboard[%s] connector JSON bytes=%d", countyID, len(configJSON))

	now := time.Now()
	store := registry.NewStore(pool)
	log.Printf("onboard[%s] upserting registry row", countyID)
	err = store.Upsert(ctx, &registry.CountyConnector{
		CountyID:        countyID,
		CountyName:      countyName,
		State:           state,
		URL:             url,
		SourceType:      sourceType,
		ConnectorConfig: configJSON,
		Status:          "active",
		LastGeneratedAt: &now,
	})
	if err != nil {
		return err
	}
	log.Printf("onboard[%s] complete", countyID)
	return nil
}
