package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

func splitComma(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

type Config struct {
	DatabaseURL          string
	AnthropicAPIKey      string
	AnthropicModel       string
	Env                  string
	MaxConcurrent        int
	HTTPAddr             string
	RunNowOnStartup      bool
	JWTSecret            string
	CORSAllowedOrigins   []string
}

func Load() *Config {
	_ = godotenv.Load()

	maxConc := 5
	if v := getEnv("MAX_CONCURRENT", ""); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxConc = n
		}
	}

	httpAddr := resolveHTTPAddr()

	corsOrigins := splitComma(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173"))

	return &Config{
		DatabaseURL:        mustEnv("DATABASE_URL"),
		AnthropicAPIKey:    mustEnv("ANTHROPIC_API_KEY"),
		AnthropicModel:     getEnv("ANTHROPIC_MODEL", "claude-sonnet-4-20250514"),
		Env:                getEnv("APP_ENV", "development"),
		MaxConcurrent:      maxConc,
		HTTPAddr:           httpAddr,
		RunNowOnStartup:    getEnv("GLASSBASE_RUN_NOW", "") == "true",
		JWTSecret:          mustEnv("JWT_SECRET"),
		CORSAllowedOrigins: corsOrigins,
	}
}

// resolveHTTPAddr picks the listen address for the admin HTTP server.
// Railway (and similar) set PORT; optional HTTP_ADDR overrides (e.g. :8080 for local dev).
func resolveHTTPAddr() string {
	if v := getEnv("HTTP_ADDR", ""); v != "" {
		return v
	}
	if p := getEnv("PORT", ""); p != "" {
		return ":" + p
	}
	return ":8080"
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("missing required env var: " + key)
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
