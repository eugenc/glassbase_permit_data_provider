package main

import (
	"flag"
	"log"
	"strings"

	"github.com/echayko/glassbase_permit_data_provider/config"
	"github.com/echayko/glassbase_permit_data_provider/internal/db"
)

func main() {
	urlFlag := flag.String("url", "", "postgres URL for this run only (highest precedence)")
	envFlag := flag.String("env", "", "optional: dev|prod → DATABASE_URL_DEV / DATABASE_URL_PROD (skipped when DATABASE_URL/DATABASE_PUBLIC_URL take precedence)")
	flag.Parse()

	config.LoadDotenv()

	databaseURL := strings.TrimSpace(*urlFlag)
	if databaseURL != "" {
		log.Println("migrate: using -url")
	} else {
		databaseURL = config.PickDatabaseURL(strings.TrimSpace(*envFlag))
		if databaseURL != "" && strings.TrimSpace(*envFlag) != "" && strings.TrimSpace(*urlFlag) == "" {
			log.Printf("migrate: resolved with -env=%s", strings.TrimSpace(*envFlag))
		}
	}
	if databaseURL == "" {
		log.Fatal(`migrate: missing URL — use -url, or set DATABASE_URL, or DATABASE_URL_DEV + DATABASE_URL_PROD (+ optional -env dev|prod)`)
	}

	if err := db.RunMigrations(databaseURL); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("migrations applied successfully")
}
