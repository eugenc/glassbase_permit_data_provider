package config

import (
	"os"
	"path/filepath"
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
	DatabaseURL           string
	AnthropicAPIKey       string
	AnthropicModel        string
	Env                   string
	MaxConcurrent         int
	HTTPAddr              string
	RunNowOnStartup       bool
	JWTSecret             string
	CORSAllowedOrigins    []string
	UpsertBatchMaxRows    int
	UpsertMaxBindParams   int
}

// LoadDotenv loads the nearest .env walking up from the current working directory (then cwd only).
// This matches running binaries from subdirs or outside the repo root while keeping .env at project root.
func LoadDotenv() {
	cwd, err := os.Getwd()
	if err != nil {
		_ = godotenv.Load()
		return
	}
	dir := cwd
	for range 12 {
		p := filepath.Join(dir, ".env")
		if st, statErr := os.Stat(p); statErr == nil && !st.IsDir() {
			_ = godotenv.Load(p)
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	_ = godotenv.Load(filepath.Join(cwd, ".env"))
}

func Load() *Config {
	LoadDotenv()

	maxConc := 5
	if v := getEnv("MAX_CONCURRENT", ""); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxConc = n
		}
	}

	httpAddr := resolveHTTPAddr()

	corsOrigins := splitComma(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173"))

	upsertBatch := clampInt(parseIntOptional(getEnv("GLASSBASE_UPSERT_BATCH_MAX_ROWS", "2000"), 2000), 1, 20000)
	upsertParams := clampInt(parseIntOptional(getEnv("GLASSBASE_UPSERT_MAX_BIND_PARAMS", "62000"), 62000), 1024, 65500)

	dbURL := PickDatabaseURL("")
	if dbURL == "" {
		panic("missing database URL: set DATABASE_URL or DATABASE_URL_DEV + DATABASE_URL_PROD (see .env.example)")
	}

	return &Config{
		DatabaseURL:        dbURL,
		AnthropicAPIKey:    mustEnv("ANTHROPIC_API_KEY"),
		AnthropicModel:     getEnv("ANTHROPIC_MODEL", "claude-sonnet-4-20250514"),
		Env:                getEnv("APP_ENV", "development"),
		MaxConcurrent:      maxConc,
		HTTPAddr:           httpAddr,
		RunNowOnStartup:    getEnv("GLASSBASE_RUN_NOW", "") == "true",
		JWTSecret:          mustEnv("JWT_SECRET"),
		CORSAllowedOrigins: corsOrigins,
		UpsertBatchMaxRows:  upsertBatch,
		UpsertMaxBindParams: upsertParams,
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
	v := strings.TrimSpace(os.Getenv(key))
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

func parseIntOptional(s string, fallback int) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func clampInt(n, min, max int) int {
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}
