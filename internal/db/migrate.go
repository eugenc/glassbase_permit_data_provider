package db

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// resolveMigrationsDir returns an absolute path to the migrations folder.
// If MIGRATIONS_PATH is set, it wins. Otherwise we walk up from the working directory
// until we find a "migrations" directory (same idea as config.LoadDotenv).
func resolveMigrationsDir() (string, error) {
	if v := strings.TrimSpace(os.Getenv("MIGRATIONS_PATH")); v != "" {
		return filepath.Abs(v)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return filepath.Abs("migrations")
	}
	dir := cwd
	for range 12 {
		cand := filepath.Join(dir, "migrations")
		if st, statErr := os.Stat(cand); statErr == nil && st.IsDir() {
			return filepath.Abs(cand)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return filepath.Abs("migrations")
}

func RunMigrations(databaseURL string) error {
	absDir, err := resolveMigrationsDir()
	if err != nil {
		return err
	}
	sourceURL := "file://" + filepath.ToSlash(absDir)

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"postgres",
		driver,
	)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
