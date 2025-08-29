package model

import "time"

type UploadLink struct {
	ID             int64     `db:"id"`
	Name           string    `db:"name"`
	HashedPassword string    `db:"password"`
	CreatedAt      time.Time `db:"created_at"`
	ExpiresAt      time.Time `db:"expires_at"`
	LinkToken      string    `db:"link_token"`
}
