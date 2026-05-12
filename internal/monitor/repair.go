package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/echayko/glassbase_permit_data_provider/internal/fetcher"
	"github.com/echayko/glassbase_permit_data_provider/internal/generator"
	"github.com/echayko/glassbase_permit_data_provider/internal/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repairer struct {
	pool      *pgxpool.Pool
	generator *generator.Generator
}

func NewRepairer(pool *pgxpool.Pool, apiKey, model string) *Repairer {
	return &Repairer{
		pool: pool,
		generator: &generator.Generator{
			APIKey: apiKey,
			Model:  model,
		},
	}
}

type RepairResult struct {
	CountyID string
	Success  bool
	Reason   string
}

// RepairCounty re-fetches the county page and regenerates the connector config via Claude.
func (r *Repairer) RepairCounty(ctx context.Context, countyID string) RepairResult {
	result := RepairResult{CountyID: countyID}

	store := registry.NewStore(r.pool)
	county, err := store.GetByCountyID(ctx, countyID)
	if err != nil {
		result.Reason = fmt.Sprintf("load county: %v", err)
		return result
	}
	if county == nil {
		result.Reason = "county not found"
		return result
	}

	log.Printf("[%s] repair: re-detecting source type...", countyID)

	sourceType, err := fetcher.DetectSourceType(ctx, county.URL)
	if err != nil {
		result.Reason = fmt.Sprintf("detect: %v", err)
		return result
	}

	log.Printf("[%s] repair: re-fetching page...", countyID)
	f := fetcher.New(sourceType)
	fetchResult, err := f.Fetch(ctx, county.URL)
	if err != nil {
		result.Reason = fmt.Sprintf("fetch: %v", err)
		return result
	}

	log.Printf("[%s] repair: re-generating connector config via Claude...", countyID)
	var networkSummary []string
	for _, nc := range fetchResult.NetworkCalls {
		max := 200
		if len(nc.Response) < max {
			max = len(nc.Response)
		}
		suffix := ""
		if max > 0 {
			suffix = nc.Response[:max]
		}
		networkSummary = append(networkSummary, nc.URL+" → "+suffix)
	}

	prompt := generator.BuildPrompt(county.URL, sourceType, fetchResult.Body, networkSummary)
	cfg, err := r.generator.GenerateConnectorConfig(ctx, prompt)
	if err != nil {
		result.Reason = fmt.Sprintf("generate: %v", err)
		return result
	}

	if cfg.SourceType != "" {
		sourceType = cfg.SourceType
	}

	configJSON, err := json.Marshal(cfg)
	if err != nil {
		result.Reason = fmt.Sprintf("marshal: %v", err)
		return result
	}

	now := time.Now()
	err = store.Upsert(ctx, &registry.CountyConnector{
		CountyID:        county.CountyID,
		CountyName:      county.CountyName,
		State:           county.State,
		URL:             county.URL,
		SourceType:      sourceType,
		ConnectorConfig: configJSON,
		Status:          "active",
		LastGeneratedAt: &now,
	})
	if err != nil {
		result.Reason = fmt.Sprintf("store: %v", err)
		return result
	}

	result.Success = true
	log.Printf("[%s] repair: connector regenerated and re-activated", countyID)
	return result
}

// RepairBroken finds all broken counties and attempts to repair each one.
func (r *Repairer) RepairBroken(ctx context.Context) {
	store := registry.NewStore(r.pool)
	counties, err := store.GetByStatus(ctx, "broken")
	if err != nil {
		log.Printf("repair: load broken: %v", err)
		return
	}

	if len(counties) == 0 {
		log.Println("repair: no broken counties found")
		return
	}

	log.Printf("repair: attempting to repair %d broken counties", len(counties))
	for _, county := range counties {
		res := r.RepairCounty(ctx, county.CountyID)
		if !res.Success {
			log.Printf("repair: [%s] failed — %s", county.CountyID, res.Reason)
		}
	}
}
