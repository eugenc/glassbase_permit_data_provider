package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/echayko/glassbase_permit_data_provider/config"
	"github.com/echayko/glassbase_permit_data_provider/internal/db"
	"github.com/echayko/glassbase_permit_data_provider/internal/registry"
)

func main() {
	if os.Getenv("ALLOW_PROD_DB") != "true" {
		log.Fatal("Set ALLOW_PROD_DB=true explicitly to query broken counties via this CLI")
	}

	cfg := config.Load()
	ctx := context.Background()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	store := registry.NewStore(pool)
	counties, err := store.GetByStatus(ctx, "broken")
	if err != nil {
		log.Fatalf("query: %v", err)
	}

	ids := make([]string, len(counties))
	for i, c := range counties {
		ids[i] = c.CountyID
	}

	out, err := json.Marshal(ids)
	if err != nil {
		log.Fatalf("json: %v", err)
	}
	fmt.Println(string(out))
}
