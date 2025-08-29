package repository

import (
	"context"
	"database/sql"
	"github.com/NiClassic/go-cloud/internal/model"
)

type UploadLinkSessionRepository struct {
	db *sql.DB
}

func NewUploadLinkSessionRepository(db *sql.DB) *UploadLinkSessionRepository {
	return &UploadLinkSessionRepository{db: db}
}

func (u *UploadLinkSessionRepository) Insert(
	ctx context.Context,
	userID, uploadLinkID int64, sessionToken string) (int64, error) {
	query := `INSERT INTO upload_link_sessions (user_id, upload_link_id, session_token) VALUES (?, ?, ?)`
	result, err := u.db.ExecContext(ctx, query, userID, uploadLinkID, sessionToken)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (u *UploadLinkSessionRepository) GetByToken(
	ctx context.Context,
	token string) (*model.UploadLinkSession, error) {
	query := `SELECT id, user_id, upload_link_id, created_at, valid, session_token FROM upload_link_sessions WHERE session_token = ?`

	row := u.db.QueryRowContext(ctx, query, token)

	var uploadLinkSession model.UploadLinkSession
	err := row.Scan(
		&uploadLinkSession.ID,
		&uploadLinkSession.UserID,
		&uploadLinkSession.UploadLinkID,
		&uploadLinkSession.CreatedAt,
		&uploadLinkSession.Valid,
		&uploadLinkSession.SessionToken,
	)
	return &uploadLinkSession, err
}

func (u *UploadLinkSessionRepository) Invalidate(
	ctx context.Context, token string) error {
	query := `UPDATE upload_link_sessions SET valid = 0 WHERE session_token = ?`

	_, err := u.db.ExecContext(ctx, query, token)
	return err
}
