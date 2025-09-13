package repository

import (
	"context"
	"database/sql"

	"github.com/NiClassic/go-cloud/internal/model"
)

type UploadLinkSessionRepository struct{ baseRepo }

func NewUploadLinkSessionRepository(db *sql.DB) *UploadLinkSessionRepository {
	return &UploadLinkSessionRepository{newBaseRepo(db)}
}

func (r *UploadLinkSessionRepository) Insert(
	ctx context.Context,
	userID, uploadLinkID int64,
	sessionToken string,
) (int64, error) {
	const q = `INSERT INTO upload_link_sessions (user_id, upload_link_id, session_token) VALUES (?, ?, ?)`
	res, err := r.db.ExecContext(ctx, q, userID, uploadLinkID, sessionToken)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *UploadLinkSessionRepository) GetByToken(
	ctx context.Context,
	token string,
) (*model.UploadLinkSession, error) {
	const q = `
		SELECT id, user_id, upload_link_id, created_at, valid, session_token
		FROM upload_link_sessions
		WHERE session_token = ?`
	var uls model.UploadLinkSession
	if err := r.db.QueryRowContext(ctx, q, token).Scan(
		&uls.ID, &uls.UserID, &uls.UploadLinkID, &uls.CreatedAt, &uls.Valid, &uls.SessionToken,
	); err != nil {
		return nil, err
	}
	return &uls, nil
}

func (r *UploadLinkSessionRepository) Invalidate(ctx context.Context, token string) error {
	const q = `UPDATE upload_link_sessions SET valid = 0 WHERE session_token = ?`
	_, err := r.db.ExecContext(ctx, q, token)
	return err
}
