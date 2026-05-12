package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/echayko/glassbase_permit_data_provider/config"
	"github.com/echayko/glassbase_permit_data_provider/internal/db"
	"github.com/echayko/glassbase_permit_data_provider/internal/fetcher"
	"github.com/echayko/glassbase_permit_data_provider/internal/generator"
	"github.com/echayko/glassbase_permit_data_provider/internal/registry"
)

func main() {
	csvPath := "counties.csv"
	if len(os.Args) > 1 {
		csvPath = os.Args[1]
	}

	f, err := os.Open(csvPath)
	if err != nil {
		log.Fatalf("open csv: %v", err)
	}
	defer f.Close()

	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		log.Fatalf("read csv: %v", err)
	}

	cfg := config.Load()
	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	store := registry.NewStore(pool)
	gen := &generator.Generator{APIKey: cfg.AnthropicAPIKey, Model: cfg.AnthropicModel}

	sem := make(chan struct{}, 3)
	var wg sync.WaitGroup

	for _, row := range records[1:] {
		row := row
		if len(row) < 4 {
			continue
		}
		countyID, name, state, url := row[0], row[1], row[2], row[3]

		wg.Add(1)
		sem <- struct{}{}

		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			log.Printf("[%s] onboarding...", countyID)

			sourceType, err := fetcher.DetectSourceType(ctx, url)
			if err != nil {
				log.Printf("[%s] detect failed: %v", countyID, err)
				return
			}

			fetch := fetcher.New(sourceType)
			result, err := fetch.Fetch(ctx, url)
			if err != nil {
				log.Printf("[%s] fetch failed: %v", countyID, err)
				return
			}

			var networkSummary []string
			for _, nc := range result.NetworkCalls {
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

			prompt := generator.BuildPrompt(url, sourceType, result.Body, networkSummary)
			connConfig, err := gen.GenerateConnectorConfig(ctx, prompt)
			if err != nil {
				log.Printf("[%s] generate failed: %v", countyID, err)
				return
			}

			if connConfig.SourceType != "" {
				sourceType = connConfig.SourceType
			}

			configJSON, err := json.Marshal(connConfig)
			if err != nil {
				log.Printf("[%s] marshal failed: %v", countyID, err)
				return
			}
			now := time.Now()
			err = store.Upsert(ctx, &registry.CountyConnector{
				CountyID:        countyID,
				CountyName:      name,
				State:           state,
				URL:             url,
				SourceType:      sourceType,
				ConnectorConfig: configJSON,
				Status:          "active",
				LastGeneratedAt: &now,
			})
			if err != nil {
				log.Printf("[%s] store failed: %v", countyID, err)
				return
			}

			log.Printf("[%s] onboarded (%s)", countyID, sourceType)
		}()
	}

	wg.Wait()
	log.Println("bulk onboard complete")
}
