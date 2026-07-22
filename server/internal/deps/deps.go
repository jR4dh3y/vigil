// Package deps pins arch.md §16 libraries in go.mod via blank imports.
// Real usage lives in domain packages; this file only prevents go mod tidy from dropping them.
package deps

import (
	_ "github.com/alexedwards/argon2id"
	_ "github.com/coder/websocket"
	_ "github.com/go-chi/chi/v5"
	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/knadh/koanf/parsers/yaml"
	_ "github.com/knadh/koanf/providers/env"
	_ "github.com/knadh/koanf/providers/file"
	_ "github.com/knadh/koanf/v2"
	_ "github.com/minio/minio-go/v7"
	_ "github.com/oapi-codegen/runtime"
	_ "github.com/robfig/cron/v3"
	_ "github.com/shirou/gopsutil/v4/disk"
	_ "github.com/use-go/onvif"
	_ "maragu.dev/goqite"
	_ "modernc.org/sqlite"
)
