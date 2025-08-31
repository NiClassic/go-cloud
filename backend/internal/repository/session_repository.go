package repository

import (
	"context"
	"database/sql"

	"github.com/NiClassic/go-cloud/internal/model"
)

type SessionRepository struct{ baseRepo }

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{newBaseRepo(db)}
}

func (r *SessionRepository) Insert(
	ctx context.Context,
	userID int64,
	sessionToken string,
) (int64, error) {
	const q = `INSERT INTO sessions (user_id, session_token) VALUES (?, ?)`
	res, err := r.db.ExecContext(ctx, q, userID, sessionToken)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *SessionRepository) GetByToken(
	ctx context.Context,
	token string,
) (*model.Session, error) {
	const q = `
		SELECT id, user_id, created_at, valid, session_token
		FROM sessions
		WHERE session_token = ?`
	var s model.Session
	if err := r.db.QueryRowContext(ctx, q, token).Scan(
		&s.ID, &s.UserID, &s.CreatedAt, &s.Valid, &s.SessionToken,
	); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SessionRepository) Invalidate(ctx context.Context, token string) error {
	const q = `UPDATE sessions SET valid = 0 WHERE session_token = ?`
	_, err := r.db.ExecContext(ctx, q, token)
	return err
}

func (r *SessionRepository) DeleteByUser(ctx context.Context, userID int64) error {
	const q = `DELETE FROM sessions WHERE user_id = ?`
	_, err := r.db.ExecContext(ctx, q, userID)
	return err
}
