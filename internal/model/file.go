package model

import (
	"database/sql"
	"time"
)

type File struct {
	ID        int64         `db:"id"`
	UserID    int64         `db:"user_id"`
	Name      string        `db:"name"`
	Size      int64         `db:"size"`
	MimeType  string        `db:"mime_type"`
	CreatedAt time.Time     `db:"created_at"`
	Location  string        `db:"location"`
	Hash      string        `db:"hash"`
	FolderID  sql.NullInt64 `db:"folder_id"`
}
