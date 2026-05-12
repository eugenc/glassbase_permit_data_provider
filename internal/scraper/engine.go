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
	pool                *pgxpool.Pool
	upsertBatchMaxRows  int
	upsertMaxBindParams int
}

func NewEngine(pool *pgxpool.Pool, upsertBatchMaxRows, upsertMaxBindParams int) *Engine {
	if upsertBatchMaxRows <= 0 {
		upsertBatchMaxRows = 2000
	}
	if upsertMaxBindParams <= 0 {
		upsertMaxBindParams = 62000
	}
	return &Engine{
		pool:                pool,
		upsertBatchMaxRows:  upsertBatchMaxRows,
		upsertMaxBindParams: upsertMaxBindParams,
	}
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

	var totalFound, totalInserted int

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
		case "csv":
			records, err = ExtractFromCSV(page.Body, &config)
		default:
			records, err = ExtractFromHTML(page.Body, &config)
		}
		if err != nil {
			log.Printf("[%s] extract page %d: %v", county.CountyID, page.PageNum, err)
			if config.Pagination.Type == "none" {
				break
			}
			// API errors (e.g. Tyler Success=false after deep pagination) will repeat every page; stop.
			if sourceType == "api" || sourceType == "csv" {
				break
			}
			continue
		}

		if len(records) == 0 {
			break
		}

		// Full HTTP response is already in memory (see fetcher ReadAll); we only persist after parse succeeds.
		totalFound += len(records)
		inserted, err := e.upsertRecords(ctx, county.CountyID, records, &config)
		if err != nil {
			result.Error = fmt.Errorf("upsert: %w", err)
			return result, result.Error
		}
		totalInserted += inserted
		log.Printf("[%s] Page %d: %d records (%d upsert rows)", county.CountyID, page.PageNum, len(records), inserted)
	}

	result.RecordsFound = totalFound
	result.RecordsInserted = totalInserted
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
	if (st == "api" || st == "csv") && cfg.API != nil {
		return fetcher.NewWithOptions(fetcher.Options{
			SourceType: st,
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

func insertColumnNames(config *generator.ConnectorConfig) []string {
	cols := []string{"permit_number", "scraped_at", "raw_data"}
	seenCol := map[string]struct{}{"permit_number": {}, "scraped_at": {}, "raw_data": {}}
	for _, f := range config.Extraction.Fields {
		col := sanitizeColName(f.Name)
		if isFixedPermitColumn(col) {
			continue
		}
		if _, dup := seenCol[col]; dup {
			continue
		}
		seenCol[col] = struct{}{}
		cols = append(cols, col)
	}
	return cols
}

func buildRowValues(record PermitRecord, config *generator.ConnectorConfig, scrapedAt time.Time) ([]interface{}, bool, error) {
	permitNumber := record[config.Dedup.UniqueField]
	if permitNumber == "" {
		return nil, false, nil
	}
	rawJSON, err := json.Marshal(record)
	if err != nil {
		return nil, false, err
	}
	vals := []interface{}{permitNumber, scrapedAt, rawJSON}
	seenCol := map[string]struct{}{"permit_number": {}, "scraped_at": {}, "raw_data": {}}
	for _, f := range config.Extraction.Fields {
		col := sanitizeColName(f.Name)
		if isFixedPermitColumn(col) {
			continue
		}
		if _, dup := seenCol[col]; dup {
			continue
		}
		seenCol[col] = struct{}{}
		vals = append(vals, record[f.Name])
	}
	return vals, true, nil
}

func (e *Engine) upsertRecords(ctx context.Context, countyID string, records []PermitRecord, config *generator.ConnectorConfig) (int, error) {
	if len(records) == 0 {
		return 0, nil
	}

	tableName := tableNameFor(countyID)
	cols := insertColumnNames(config)
	nCols := len(cols)
	if nCols == 0 {
		return 0, fmt.Errorf("no columns for upsert")
	}

	batchMax := e.upsertBatchMaxRows
	if config.RateLimit.UpsertBatchMaxRows > 0 {
		batchMax = config.RateLimit.UpsertBatchMaxRows
	}

	batchRows := e.upsertMaxBindParams / nCols
	if batchRows < 1 {
		batchRows = 1
	}
	if batchRows > batchMax {
		batchRows = batchMax
	}

	inserted := 0
	scrapedAt := time.Now()

	for start := 0; start < len(records); start += batchRows {
		end := start + batchRows
		if end > len(records) {
			end = len(records)
		}
		chunk := records[start:end]

		var rowSQL []string
		args := make([]interface{}, 0, len(chunk)*nCols)
		argPos := 1
		inBatch := make(map[string]struct{}, len(chunk))

		for _, record := range chunk {
			pn := record[config.Dedup.UniqueField]
			if pn != "" {
				if _, dup := inBatch[pn]; dup {
					continue
				}
				inBatch[pn] = struct{}{}
			}

			vals, ok, err := buildRowValues(record, config, scrapedAt)
			if err != nil {
				return inserted, err
			}
			if !ok {
				continue
			}
			if len(vals) != nCols {
				return inserted, fmt.Errorf("column/value mismatch: %d cols vs %d vals", nCols, len(vals))
			}
			ph := make([]string, len(vals))
			for i := range vals {
				ph[i] = fmt.Sprintf("$%d", argPos)
				argPos++
			}
			rowSQL = append(rowSQL, "("+strings.Join(ph, ", ")+")")
			args = append(args, vals...)
		}

		if len(rowSQL) == 0 {
			continue
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
			VALUES %s
			ON CONFLICT (permit_number) DO UPDATE SET %s`,
			tableName,
			strings.Join(cols, ", "),
			strings.Join(rowSQL, ", "),
			strings.Join(updateParts, ", "),
		)

		tag, err := e.pool.Exec(ctx, q, args...)
		if err != nil {
			return inserted, fmt.Errorf("batch upsert (%d rows): %w", len(rowSQL), err)
		}
		affected := int(tag.RowsAffected())
		inserted += affected
		log.Printf("[%s] db: committed batch of %d permit row(s) — postgres rows_affected=%d (insert + conflict updates)",
			countyID, len(rowSQL), affected)
	}

	if inserted > 0 {
		log.Printf("[%s] db: upsert complete for this page chunk — total rows_affected=%d", countyID, inserted)
	}

	return inserted, nil
}
