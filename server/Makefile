.PHONY: run build generate tidy sqlc oapi

run:
	go run ./cmd/nvrd

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -o bin/nvrd ./cmd/nvrd

# Requires: go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
#           go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
generate: oapi sqlc

oapi:
	oapi-codegen -config api/oapi-codegen.yaml api/openapi.yaml

sqlc:
	sqlc generate

tidy:
	go mod tidy
