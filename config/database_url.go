package config

import (
	"os"
	"strings"
)

// PickDatabaseURL chooses the Postgres URL after godotenv (or other) has loaded env.
//
// Precedence when migrateEnv == "":
//  1. DATABASE_URL — single URL for Railway / docker-compose / CI (backward compatible).
//  2. DATABASE_PUBLIC_URL — optional public proxy URL (e.g. Railway from your laptop).
//  3. DATABASE_URL_PROD when isProductionLike(), else DATABASE_URL_DEV.
//
// When migrateEnv is "dev"|"development" or "prod"|"production", step 3 uses that slot
// (skips DATABASE_URL / DATABASE_PUBLIC_URL so you can target dev/prod while those are set elsewhere).
func PickDatabaseURL(migrateEnv string) string {
	me := normalizeMigrateEnv(migrateEnv)
	explicitBranch := me == "dev" || me == "prod"

	if !explicitBranch {
		if v := trimEnvVar("DATABASE_URL"); v != "" {
			return v
		}
		if v := trimEnvVar("DATABASE_PUBLIC_URL"); v != "" {
			return v
		}
		if isProductionLike() {
			if v := trimEnvVar("DATABASE_URL_PROD"); v != "" {
				return v
			}
		}
		if v := trimEnvVar("DATABASE_URL_DEV"); v != "" {
			return v
		}
		return ""
	}

	switch me {
	case "prod":
		return trimEnvVar("DATABASE_URL_PROD")
	default:
		return trimEnvVar("DATABASE_URL_DEV")
	}
}

func normalizeMigrateEnv(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "development":
		return "dev"
	case "production":
		return "prod"
	default:
		return s
	}
}

func isProductionLike() bool {
	app := strings.TrimSpace(strings.ToLower(os.Getenv("APP_ENV")))
	if app == "production" || app == "prod" {
		return true
	}
	if strings.TrimSpace(strings.ToLower(os.Getenv("RAILWAY_ENVIRONMENT"))) == "production" {
		return true
	}
	return false
}

func trimEnvVar(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}
