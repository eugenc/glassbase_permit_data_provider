package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/echayko/glassbase_permit_data_provider/internal/db"
)

func main() {
	_ = godotenv.Load()

	// Railway: DATABASE_URL uses an internal host on the platform; DATABASE_PUBLIC_URL works from your machine.
	databaseURL := os.Getenv("DATABASE_PUBLIC_URL")
	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
	}
	if databaseURL == "" {
		log.Fatal("missing DATABASE_URL or DATABASE_PUBLIC_URL")
	}

	if err := db.RunMigrations(databaseURL); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("migrations applied successfully")
}
