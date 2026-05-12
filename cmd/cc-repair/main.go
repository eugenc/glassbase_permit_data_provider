package main

import (
	"context"
	"flag"
	"log"

	"github.com/echayko/glassbase_permit_data_provider/config"
	"github.com/echayko/glassbase_permit_data_provider/internal/db"
	"github.com/echayko/glassbase_permit_data_provider/internal/repair"
	"github.com/echayko/glassbase_permit_data_provider/internal/registry"
)

func main() {
	countyID := flag.String("county", "", "County ID (required)")
	trigger := flag.String("trigger", "manual", `Trigger label: manual | zero_records | health_probe`)
	flag.Parse()

	if *countyID == "" {
		log.Fatal("--county is required")
	}

	cfg := config.Load()
	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	store := registry.NewStore(pool)
	c, err := store.GetByCountyID(ctx, *countyID)
	if err != nil {
		log.Fatalf("load county: %v", err)
	}
	if c == nil {
		log.Fatal("county not found")
	}

	runner := repair.NewRunner(pool)
	out, err := runner.RepairCounty(ctx, c, *trigger)
	if out == nil {
		log.Fatalf("repair: %v", err)
	}
	if err != nil {
		log.Printf("[repair-ai] CLI cc-repair: repair ended with error: %v", err)
	}
	log.Printf("[repair-ai] CLI cc-repair: success=%v commit=%s pr=%s\n%s", out.Success, out.CommitSHA, out.PRUrl, out.Output)
}
