package main

import (
	"context"
	"flag"
	"log"

	"github.com/echayko/glassbase_permit_data_provider/config"
	"github.com/echayko/glassbase_permit_data_provider/internal/db"
	"github.com/echayko/glassbase_permit_data_provider/internal/onboard"
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

	log.Printf("Onboarding %s...", *countyID)
	if err := onboard.Run(ctx, pool, cfg, *countyID, *name, *state, *url); err != nil {
		log.Fatalf("onboard: %v", err)
	}

	log.Printf("County %s onboarded successfully", *countyID)
}
