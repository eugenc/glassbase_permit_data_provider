.PHONY: scrape-one diagnose migrate test vet cc-repair build web-build build-all build-all-run run help

# --- Build + run ---
# Use:  make build-all   then   ./bin/glassbase
# Or:   make build-all-run     (same, in one make; Ctrl+C stops the server)
# Do not run: make build-all # ./bin/glassbase   — shells differ; use && or two lines.

help:
	@echo "make web-build       - npm ci + vite build → internal/static/web"
	@echo "make build           - go build → bin/glassbase"
	@echo "make build-all       - web-build + build"
	@echo "make build-all-run   - build-all then ./bin/glassbase"
	@echo "make run             - build then ./bin/glassbase (binary only)"

COUNTY ?= 

web-build:
	npm --prefix frontend ci
	npm --prefix frontend run build

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -o bin/glassbase ./cmd/server

build-all: web-build build

build-all-run: build-all
	./bin/glassbase

run: build
	./bin/glassbase

# Single-county scrape (requires COUNTY=slug)
scrape-one:
	go run ./cmd/scrape-one --county=$(COUNTY)

diagnose:
	go run ./cmd/diagnose --county=$(COUNTY)

migrate:
	go run ./cmd/migrate

test:
	go test ./...

vet:
	go vet ./...

cc-repair:
	go run ./cmd/cc-repair --county=$(COUNTY)
