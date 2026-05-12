package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/echayko/glassbase_permit_data_provider/config"
	"github.com/echayko/glassbase_permit_data_provider/internal/db"
	"github.com/echayko/glassbase_permit_data_provider/internal/fetcher"
	"github.com/echayko/glassbase_permit_data_provider/internal/generator"
	"github.com/echayko/glassbase_permit_data_provider/internal/registry"
)

func main() {
	county := flag.String("county", "", "County ID to diagnose (required)")
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
	target, err := store.GetByCountyID(ctx, *county)
	if err != nil {
		log.Fatalf("load county: %v", err)
	}
	if target == nil {
		log.Fatalf("county %q not found", *county)
	}

	fmt.Println("=== STORED CONNECTOR CONFIG ===")
	var confPretty map[string]interface{}
	if json.Unmarshal(target.ConnectorConfig, &confPretty) == nil {
		b, _ := json.MarshalIndent(confPretty, "", "  ")
		fmt.Println(string(b))
	} else {
		fmt.Println(string(target.ConnectorConfig))
	}

	var typed generator.ConnectorConfig
	_ = json.Unmarshal(target.ConnectorConfig, &typed)

	fmt.Println("\n=== LIVE PAGE FETCH ===")
	sourceType, err := fetcher.DetectSourceType(ctx, target.URL)
	if err != nil {
		log.Fatalf("detect: %v", err)
	}
	fmt.Printf("Detected source type: %s\n", sourceType)

	f := fetcher.New(sourceType)
	result, err := f.Fetch(ctx, target.URL, nil)
	if err != nil {
		log.Fatalf("fetch: %v", err)
	}
	fmt.Printf("Body length: %d bytes\n", len(result.Body))
	fmt.Printf("Status code: %d\n", result.StatusCode)

	if len(result.NetworkCalls) > 0 {
		fmt.Printf("\nIntercepted %d network calls:\n", len(result.NetworkCalls))
		for i, nc := range result.NetworkCalls {
			if i >= 5 {
				fmt.Printf("  ... and %d more\n", len(result.NetworkCalls)-5)
				break
			}
			fmt.Printf("  %s %s\n", nc.Method, nc.URL)
		}
	}

	fmt.Println("\n=== SELECTOR PROBE ===")
	sel := typed.Extraction.RecordSelector
	if sel == "" && confPretty != nil {
		if ext, ok := confPretty["extraction"].(map[string]interface{}); ok {
			if s, ok := ext["record_selector"].(string); ok {
				sel = s
			}
		}
	}
	if sel != "" {
		fmt.Printf("Stored record_selector: %q\n", sel)
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(result.Body))
		if err != nil {
			fmt.Printf("goquery parse: %v\n", err)
		} else {
			count := doc.Find(sel).Length()
			fmt.Printf("Matches on live page: %d\n", count)
			if count == 0 {
				fmt.Println("SELECTOR LIKELY MISMATCH — common cause of 0 records")
				fmt.Println("\nCandidate selectors (non-zero matches):")
				for _, candidate := range []string{
					"table tr", "tbody tr", ".permit-row", ".result-row",
					".record", ".permit", "[class*='permit']", "[class*='result']",
					"[class*='row']", "li", ".list-item",
				} {
					n := doc.Find(candidate).Length()
					if n > 0 {
						fmt.Printf("  %-30s %d matches\n", candidate, n)
					}
				}
			} else {
				fmt.Println("Selector still matches DOM")
			}
		}
	} else {
		fmt.Println("(no HTML record_selector in config — api/json extraction may omit this)")
	}

	fmt.Println("\n=== PAGE BODY PREVIEW (first 2000 chars) ===")
	body := result.Body
	if len(body) > 2000 {
		body = body[:2000] + "\n... [truncated]"
	}
	fmt.Println(body)
}
