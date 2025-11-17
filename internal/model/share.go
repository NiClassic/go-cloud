package model

import (
	"database/sql"
	"time"
)

const (
	SharePermissionRead  = "read"
	SharePermissionWrite = "write"
)

type FileShare struct {
	ID           int64
	FileID       int64
	SharedWithID int64
	Permission   string
	CreatedAt    time.Time
	ExpiresAt    sql.NullTime
}
