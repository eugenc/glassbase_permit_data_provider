package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/echayko/glassbase_permit_data_provider/internal/fetcher"
	"github.com/echayko/glassbase_permit_data_provider/internal/generator"
	"github.com/echayko/glassbase_permit_data_provider/internal/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScrapeResult struct {
	CountyID        string
	RecordsFound    int
	RecordsInserted int
	Error           error
	Duration        time.Duration
}

type Engine struct {
	pool *pgxpool.Pool
}

func NewEngine(pool *pgxpool.Pool) *Engine {
	return &Engine{pool: pool}
}

func (e *Engine) ScrapeCounty(ctx context.Context, county *registry.CountyConnector) (*ScrapeResult, error) {
	start := time.Now()
	result := &ScrapeResult{CountyID: county.CountyID}

	var config generator.ConnectorConfig
	if err := json.Unmarshal(county.ConnectorConfig, &config); err != nil {
		result.Error = fmt.Errorf("invalid connector config: %w", err)
		return result, result.Error
	}

	var fieldNames []string
	for _, f := range config.Extraction.Fields {
		fieldNames = append(fieldNames, f.Name)
	}
	if err := EnsureTable(ctx, e.pool, county.CountyID, fieldNames); err != nil {
		result.Error = fmt.Errorf("ensure table: %w", err)
		return result, result.Error
	}

	f := fetcherForCounty(county, &config)
	paginator := NewPaginator(&config, county.URL, f)

	var allRecords []PermitRecord

	for {
		page, ok, err := paginator.Next(ctx)
		if err != nil {
			result.Error = fmt.Errorf("page fetch: %w", err)
			return result, result.Error
		}
		if !ok {
			break
		}

		log.Printf("[%s] Scraping page %d (%s)...", county.CountyID, page.PageNum, page.FetchURL)

		sourceType := effectiveSourceType(county, &config)
		var records []PermitRecord
		switch sourceType {
		case "api":
			records, err = ExtractFromAPI(page.Body, &config)
		default:
			records, err = ExtractFromHTML(page.Body, &config)
		}
		if err != nil {
			log.Printf("[%s] extract page %d: %v", county.CountyID, page.PageNum, err)
			if config.Pagination.Type == "none" {
				break
			}
			continue
		}

		if len(records) == 0 {
			break
		}

		allRecords = append(allRecords, records...)
		log.Printf("[%s] Page %d: %d records", county.CountyID, page.PageNum, len(records))
	}

	result.RecordsFound = len(allRecords)

	inserted, err := e.upsertRecords(ctx, county.CountyID, allRecords, &config)
	if err != nil {
		result.Error = fmt.Errorf("upsert: %w", err)
		return result, result.Error
	}

	result.RecordsInserted = inserted
	result.Duration = time.Since(start)
	return result, nil
}

func effectiveSourceType(county *registry.CountyConnector, cfg *generator.ConnectorConfig) string {
	if cfg.SourceType != "" {
		return cfg.SourceType
	}
	return county.SourceType
}

func fetcherForCounty(county *registry.CountyConnector, cfg *generator.ConnectorConfig) fetcher.Fetcher {
	st := effectiveSourceType(county, cfg)
	if st == "api" && cfg.API != nil {
		return fetcher.NewWithOptions(fetcher.Options{
			SourceType: "api",
			API: &fetcher.APIFetcher{
				Endpoint: cfg.API.Endpoint,
				Method:   cfg.API.Method,
				Headers:  cfg.API.Headers,
				Body:     cfg.API.Body,
			},
		})
	}
	return fetcher.New(st)
}

func (e *Engine) upsertRecords(ctx context.Context, countyID string, records []PermitRecord, config *generator.ConnectorConfig) (int, error) {
	tableName := tableNameFor(countyID)
	inserted := 0

	for _, record := range records {
		permitNumber := record[config.Dedup.UniqueField]
		if permitNumber == "" {
			continue
		}

		rawJSON, err := json.Marshal(record)
		if err != nil {
			return inserted, err
		}

		cols := []string{"permit_number", "scraped_at", "raw_data"}
		vals := []interface{}{permitNumber, time.Now(), rawJSON}

		for _, f := range config.Extraction.Fields {
			col := sanitizeColName(f.Name)
			cols = append(cols, col)
			vals = append(vals, record[f.Name])
		}

		placeholders := make([]string, len(vals))
		for i := range vals {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		}

		var updateParts []string
		for _, col := range cols {
			if col == "permit_number" {
				continue
			}
			updateParts = append(updateParts, fmt.Sprintf("%s = EXCLUDED.%s", col, col))
		}

		q := fmt.Sprintf(`
			INSERT INTO %s (%s)
			VALUES (%s)
			ON CONFLICT (permit_number) DO UPDATE SET %s`,
			tableName,
			strings.Join(cols, ", "),
			strings.Join(placeholders, ", "),
			strings.Join(updateParts, ", "),
		)

		tag, err := e.pool.Exec(ctx, q, vals...)
		if err != nil {
			log.Printf("upsert permit %s: %v", permitNumber, err)
			continue
		}
		if tag.RowsAffected() > 0 {
			inserted++
		}
	}

	return inserted, nil
}
