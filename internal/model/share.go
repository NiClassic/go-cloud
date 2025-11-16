package model

import "time"

const (
	ShareResourceTypeFile   = "file"
	ShareResourceTypeFolder = "folder"
	SharePermissionRead     = "read"
	SharePermissionWrite    = "write"
)

type Share struct {
	ID                 int64     `db:"id"`
	SharedByUserID     int64     `db:"shared_by_user_id"`
	SharedWithUserID   int64     `db:"shared_with_user_id"`
	ResourceType       string    `db:"resource_type"` // "file" or "folder"
	ResourceID         int64     `db:"resource_id"`
	Permission         string    `db:"permission"` // "read" or "write"
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
	SharedByUsername   string    `db:"shared_by_username"`
	SharedWithUsername string    `db:"shared_with_username"`
	ResourceName       string    `db:"resource_name"`
}
