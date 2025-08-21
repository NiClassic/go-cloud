package model

import "time"

type Session struct {
	ID           int64     `db:"id"`
	UserID       int64     `db:"user_id"`
	SessionToken string    `db:"session_token"`
	CreatedAt    time.Time `db:"created_at"`
	Valid        bool      `db:"valid"`
}
