.PHONY: run run-now build test migrate onboard bulk-onboard

build:
	go build -o bin/glassbase ./cmd/server

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
