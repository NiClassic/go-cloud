package model

import (
	"database/sql"
	"strings"
	"time"
)

type Folder struct {
	ID        int64         `db:"id"`
	UserID    int64         `db:"user_id"`
	ParentID  sql.NullInt64 `db:"parent_id"`
	Name      string        `db:"name"`
	Path      string        `db:"path"`
	CreatedAt time.Time     `db:"created_at"`
	UpdatedAt time.Time     `db:"updated_at"`
}

func (f *Folder) PathWithoutUsername(username string) string {
	return strings.TrimPrefix(f.Path, "/"+username)
}
