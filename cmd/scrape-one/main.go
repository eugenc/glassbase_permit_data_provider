package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/echayko/glassbase_permit_data_provider/config"
	"github.com/echayko/glassbase_permit_data_provider/internal/db"
	"github.com/echayko/glassbase_permit_data_provider/internal/registry"
	"github.com/echayko/glassbase_permit_data_provider/internal/scraper"
	"github.com/jackc/pgx/v5"
)

func main() {
	county := flag.String("county", "", "County ID to scrape (required)")
	verbose := flag.Bool("verbose", false, "Print sample raw_data rows from permits table")
	flag.Parse()

	if *county == "" {
		log.Fatal("--county is required")
	}

	cfg := config.Load()
	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	store := registry.NewStore(pool)
	all, err := store.GetAll(ctx)
	if err != nil {
		log.Fatalf("load counties: %v", err)
	}

	var target *registry.CountyConnector
	for i := range all {
		if all[i].CountyID == *county {
			target = &all[i]
			break
		}
	}
	if target == nil {
		log.Fatalf("county %q not found in registry", *county)
	}

	fmt.Printf("County:      %s\n", target.CountyName)
	fmt.Printf("Status:      %s\n", target.Status)
	fmt.Printf("Source Type: %s\n", target.SourceType)
	fmt.Printf("URL:         %s\n", target.URL)
	fmt.Printf("Config:      %s\n\n", string(target.ConnectorConfig))

	eng := scraper.NewEngine(pool, cfg.UpsertBatchMaxRows, cfg.UpsertMaxBindParams)
	result, err := eng.ScrapeCounty(ctx, target)
	if err != nil {
		fmt.Printf("FAILED: %v\n", err)
		log.Fatal(err)
	}

	if result.RecordsFound == 0 {
		fmt.Println("WARNING: scrape finished with 0 records")
		fmt.Println("  Selector drift or empty source is likely. Try: go run ./cmd/diagnose --county=" + *county)
	} else {
		fmt.Printf("found=%d inserted=%d duration=%s\n",
			result.RecordsFound, result.RecordsInserted, result.Duration)
	}

	if !*verbose || result.RecordsFound == 0 {
		return
	}

	ident := pgx.Identifier{scraper.PermitTableName(target.CountyID)}.Sanitize()
	q := fmt.Sprintf(`SELECT raw_data FROM %s ORDER BY scraped_at DESC LIMIT 3`, ident)
	rows, err := pool.Query(ctx, q)
	if err != nil {
		log.Printf("verbose sample query: %v", err)
		return
	}
	defer rows.Close()

	printed := 0
	for rows.Next() {
		var raw json.RawMessage
		if rows.Scan(&raw) != nil {
			continue
		}
		b, _ := json.MarshalIndent(raw, "  ", "  ")
		fmt.Printf("\nSample record raw_data:\n  %s\n", b)
		printed++
	}
	if printed == 0 {
		fmt.Println("\n(verbose: no rows returned from permits table)")
	}
}
