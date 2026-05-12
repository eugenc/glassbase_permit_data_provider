.PHONY: run run-now build build-frontend build-all test migrate onboard bulk-onboard env-railway

build:
	go build -o bin/glassbase ./cmd/server

build-frontend:
	cd frontend && npm ci && npm run build

build-all: build-frontend build

run:
	go run ./cmd/server

run-now:
	GLASSBASE_RUN_NOW=true go run ./cmd/server

migrate:
	go run ./cmd/server

test:
	go test ./...

onboard:
	go run ./cmd/onboard --url="$(URL)" --county="$(COUNTY)" --name="$(NAME)" --state="$(STATE)"

bulk-onboard:
	go run ./cmd/bulk-onboard $(CSV)

# Requires: brew install jq (or jq in PATH), railway CLI linked to project.
env-railway:
	./scripts/sync-env-from-railway.sh
