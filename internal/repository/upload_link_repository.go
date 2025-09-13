package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/NiClassic/go-cloud/internal/model"
)

type UploadLinkRepository struct{ baseRepo }

func NewUploadLinkRepository(db *sql.DB) *UploadLinkRepository {
	return &UploadLinkRepository{newBaseRepo(db)}
}

func (r *UploadLinkRepository) Insert(
	ctx context.Context,
	hashedPassword, linkToken, name string,
	expiresAt time.Time,
) (int64, error) {
	const q = `INSERT INTO upload_links (password, expires_at, link_token, name) VALUES (?, ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, q, hashedPassword, expiresAt, linkToken, name)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *UploadLinkRepository) GetByToken(
	ctx context.Context,
	linkToken string,
) (*model.UploadLink, error) {
	const q = `
		SELECT id, password, name, created_at, expires_at, link_token
		FROM upload_links
		WHERE link_token = ?`
	var ul model.UploadLink
	if err := r.db.QueryRowContext(ctx, q, linkToken).Scan(
		&ul.ID, &ul.HashedPassword, &ul.Name, &ul.CreatedAt, &ul.ExpiresAt, &ul.LinkToken,
	); err != nil {
		return nil, err
	}
	return &ul, nil
}

func (r *UploadLinkRepository) GetAll(ctx context.Context) ([]*model.UploadLink, error) {
	const q = `SELECT id, name, created_at, expires_at, link_token FROM upload_links`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer r.closeRows(rows)

	var links []*model.UploadLink
	for rows.Next() {
		var ul model.UploadLink
		if err := rows.Scan(
			&ul.ID, &ul.Name, &ul.CreatedAt, &ul.ExpiresAt, &ul.LinkToken,
		); err != nil {
			return nil, err
		}
		links = append(links, &ul)
	}
	return links, rows.Err()
}
