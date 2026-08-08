.PHONY: run build build-full build-slim generate tidy sqlc oapi

run:
	go run ./cmd/nvrd

# Full build embeds the dashboard SPA (the default artifact).
build: build-full

build-full:
	mkdir -p bin
	CGO_ENABLED=0 go build -o bin/nvrd ./cmd/nvrd

# Slim/headless build: no embedded dashboard, serves a connection page instead.
build-slim:
	mkdir -p bin
	CGO_ENABLED=0 go build -tags slim -o bin/nvrd-slim ./cmd/nvrd

# Requires: go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
#           go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
generate: oapi sqlc

oapi:
	oapi-codegen -config api/oapi-codegen.yaml api/openapi.yaml

sqlc:
	sqlc generate

tidy:
	go mod tidy