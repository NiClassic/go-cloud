package repository

import (
	"context"
	"database/sql"
	"github.com/NiClassic/go-cloud/internal/model"
	"time"
)

type FileShareRepository interface {
	Create(ctx context.Context, fs *model.FileShare) error
	GetByRecipient(ctx context.Context, userID int64) ([]model.FileShare, error)
	GetSharedFilesForRecipient(ctx context.Context, userID int64) ([]SharedFile, error)
	GetByID(ctx context.Context, fileShareID int64) (*model.FileShare, error)
	Delete(ctx context.Context, id int64) error
}

type SharedFile struct {
	ID           int64
	Name         string
	CreatedAt    time.Time
	ExpiresAt    sql.NullTime
	Size         int64
	SharedByName string
	Permission   string
}

type FileShareRepositoryImpl struct {
	db *sql.DB
}

func NewFileShareRepositoryImpl(db *sql.DB) *FileShareRepositoryImpl {
	return &FileShareRepositoryImpl{db}
}

func (r *FileShareRepositoryImpl) Create(ctx context.Context, fs *model.FileShare) error {
	const q = `INSERT INTO file_shares (file_id, shared_with_id, permission, expires_at) VALUES (?, ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, q, fs.FileID, fs.SharedWithID, fs.Permission, fs.ExpiresAt)
	if err != nil {
		return err
	}

	id, _ := res.LastInsertId()
	fs.ID = id
	return nil
}

func (r *FileShareRepositoryImpl) GetByRecipient(ctx context.Context, userID int64) ([]model.FileShare, error) {
	const q = `SELECT id, file_id, shared_with_id, permission, created_at, expires_at FROM file_shares WHERE shared_with_id = ?`
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var shares []model.FileShare
	for rows.Next() {
		var s model.FileShare
		if err = rows.Scan(&s.ID, &s.FileID, &s.SharedWithID, &s.Permission, &s.CreatedAt, &s.ExpiresAt); err != nil {
			return nil, err
		}
		shares = append(shares, s)
	}
	return shares, nil
}

func (r *FileShareRepositoryImpl) GetByID(ctx context.Context, shareID int64) (*model.FileShare, error) {
	const q = `SELECT id, file_id, shared_with_id, permission, created_at, expires_at FROM file_shares WHERE id = ?`
	row := r.db.QueryRowContext(ctx, q, shareID)
	if row.Err() != nil {
		return nil, row.Err()
	}
	var s model.FileShare
	if err := row.Scan(&s.ID, &s.FileID, &s.SharedWithID, &s.Permission, &s.CreatedAt, &s.ExpiresAt); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *FileShareRepositoryImpl) GetSharedFilesForRecipient(ctx context.Context, userID int64) ([]SharedFile, error) {
	query := `
        SELECT f.ID, f.name, f.created_at, s.expires_at, f.size, u.username, s.permission
        FROM file_shares s
        JOIN files f ON s.file_id = f.id
		JOIN users u ON u.id = f.user_id
        WHERE s.shared_with_id = ? 
          AND (s.expires_at IS NULL OR s.expires_at > datetime('now'))`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var shares []SharedFile
	for rows.Next() {
		var s SharedFile
		if err = rows.Scan(&s.ID, &s.Name, &s.CreatedAt, &s.ExpiresAt, &s.Size, &s.SharedByName, &s.Permission); err != nil {
			return nil, err
		}
		shares = append(shares, s)
	}
	return shares, nil
}

func (r *FileShareRepositoryImpl) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM file_shares WHERE id = ?`
	_, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}
	return nil
}
