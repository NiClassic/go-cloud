package model

import "time"

type UploadLinkSession struct {
	ID           int64     `db:"id"`
	UserID       int64     `db:"user_id"`
	UploadLinkID int64     `db:"upload_link_id"`
	CreatedAt    time.Time `db:"created_at"`
	Valid        bool      `db:"valid"`
	SessionToken string    `db:"session_token"`
}
