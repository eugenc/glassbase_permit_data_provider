package config

import "testing"

func TestPickDatabaseURL_precedence(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DATABASE_PUBLIC_URL", "")
	t.Setenv("DATABASE_URL_DEV", "")
	t.Setenv("DATABASE_URL_PROD", "")
	t.Setenv("APP_ENV", "development")
	t.Setenv("RAILWAY_ENVIRONMENT", "")

	const legacy = "postgres://legacy/db"
	t.Setenv("DATABASE_URL", legacy)
	if got := PickDatabaseURL(""); got != legacy {
		t.Fatalf("DATABASE_URL: got %q", got)
	}
	t.Setenv("DATABASE_URL", "")

	t.Setenv("DATABASE_PUBLIC_URL", "postgres://pub/db")
	if got := PickDatabaseURL(""); got != "postgres://pub/db" {
		t.Fatalf("DATABASE_PUBLIC_URL: got %q", got)
	}
	t.Setenv("DATABASE_PUBLIC_URL", "")

	t.Setenv("DATABASE_URL_PROD", "postgres://prod/db")
	t.Setenv("DATABASE_URL_DEV", "postgres://dev/db")
	t.Setenv("APP_ENV", "production")
	if got := PickDatabaseURL(""); got != "postgres://prod/db" {
		t.Fatalf("production: got %q", got)
	}

	t.Setenv("APP_ENV", "development")
	if got := PickDatabaseURL(""); got != "postgres://dev/db" {
		t.Fatalf("development: got %q", got)
	}

	if got := PickDatabaseURL("prod"); got != "postgres://prod/db" {
		t.Fatalf("explicit prod: got %q", got)
	}
	if got := PickDatabaseURL("dev"); got != "postgres://dev/db" {
		t.Fatalf("explicit dev: got %q", got)
	}
}

func TestPickDatabaseURL_explicitSkipsLegacy(t *testing.T) {
	const legacy = "postgres://legacy/db"
	t.Setenv("DATABASE_URL", legacy)
	t.Setenv("DATABASE_URL_DEV", "postgres://devonly/db")
	t.Setenv("DATABASE_URL_PROD", "postgres://prodonly/db")
	if got := PickDatabaseURL("prod"); got != "postgres://prodonly/db" {
		t.Fatalf("explicit prod must use DATABASE_URL_PROD, got %q", got)
	}
}
