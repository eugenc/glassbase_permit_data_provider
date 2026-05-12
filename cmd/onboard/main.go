package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"

	"github.com/echayko/glassbase_permit_data_provider/config"
	"github.com/echayko/glassbase_permit_data_provider/internal/db"
	"github.com/echayko/glassbase_permit_data_provider/internal/fetcher"
	"github.com/echayko/glassbase_permit_data_provider/internal/generator"
	"github.com/echayko/glassbase_permit_data_provider/internal/registry"
	"time"
)

func main() {
	url := flag.String("url", "", "County permit page URL (required)")
	countyID := flag.String("county", "", "County slug e.g. broward_fl (required)")
	name := flag.String("name", "", "Human name e.g. 'Broward County' (required)")
	state := flag.String("state", "", "Two-letter state code e.g. FL (required)")
	flag.Parse()

	if *url == "" || *countyID == "" || *name == "" || *state == "" {
		log.Fatal("--url, --county, --name, and --state are all required")
	}

	cfg := config.Load()
	ctx := context.Background()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	log.Printf("Detecting source type for %s...", *url)
	sourceType, err := fetcher.DetectSourceType(ctx, *url)
	if err != nil {
		log.Fatalf("detect: %v", err)
	}
	log.Printf("Source type: %s", sourceType)

	log.Printf("Fetching page...")
	f := fetcher.New(sourceType)
	result, err := f.Fetch(ctx, *url)
	if err != nil {
		log.Fatalf("fetch: %v", err)
	}
	log.Printf("Fetched %d bytes", len(result.Body))

	log.Printf("Sending to Claude for analysis...")
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

	prompt := generator.BuildPrompt(*url, sourceType, result.Body, networkSummary)
	gen := &generator.Generator{APIKey: cfg.AnthropicAPIKey, Model: cfg.AnthropicModel}
	connCfg, err := gen.GenerateConnectorConfig(ctx, prompt)
	if err != nil {
		log.Fatalf("generator: %v", err)
	}

	if connCfg.SourceType != "" {
		sourceType = connCfg.SourceType
	}

	configJSON, err := json.Marshal(connCfg)
	if err != nil {
		log.Fatalf("marshal config: %v", err)
	}
	now := time.Now()
	store := registry.NewStore(pool)
	err = store.Upsert(ctx, &registry.CountyConnector{
		CountyID:        *countyID,
		CountyName:      *name,
		State:           *state,
		URL:             *url,
		SourceType:      sourceType,
		ConnectorConfig: configJSON,
		Status:          "active",
		LastGeneratedAt: &now,
	})
	if err != nil {
		log.Fatalf("store: %v", err)
	}

	log.Printf("County %s onboarded successfully", *countyID)
	log.Printf("Connector config:\n%s", string(configJSON))
}
