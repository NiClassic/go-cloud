package repository

import (
	"context"
	"database/sql"

	"github.com/NiClassic/go-cloud/internal/model"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{
		db: db,
	}
}

// Insert creates a new session for a user and returns the session ID.
func (s *SessionRepository) Insert(
	ctx context.Context,
	userID int64,
	sessionToken string,
) (int64, error) {
	query := `
		INSERT INTO sessions (user_id, session_token)
		VALUES (?, ?)
	`
	result, err := s.db.ExecContext(ctx, query, userID, sessionToken)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetByToken retrieves a session by its token.
func (s *SessionRepository) GetByToken(
	ctx context.Context,
	token string,
) (*model.Session, error) {
	query := `
		SELECT id, user_id, created_at, valid, session_token
		FROM sessions
		WHERE session_token = ?
	`
	row := s.db.QueryRowContext(ctx, query, token)

	var session model.Session
	err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.CreatedAt,
		&session.Valid,
		&session.SessionToken,
	)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// Invalidate marks a session as invalid (logout).
func (s *SessionRepository) Invalidate(
	ctx context.Context,
	token string,
) error {
	query := `
		UPDATE sessions
		SET valid = 0
		WHERE session_token = ?
	`
	_, err := s.db.ExecContext(ctx, query, token)
	return err
}

// DeleteByUser removes all sessions for a given user (optional helper).
func (s *SessionRepository) DeleteByUser(
	ctx context.Context,
	userID int64,
) error {
	query := `DELETE FROM sessions WHERE user_id = ?`
	_, err := s.db.ExecContext(ctx, query, userID)
	return err
}
