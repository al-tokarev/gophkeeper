package migrations

// migrate -database "postgres://alexandr@localhost:5432/gophkeeper?sslmode=disable" -path ./internal/database/migrations up

import "embed"

//go:embed *.sql
var Files embed.FS
