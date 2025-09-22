package model

import "time"

type Folder struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	ParentID  int64     `db:"parent_id"`
	Name      string    `db:"name"`
	Path      string    `db:"path"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
