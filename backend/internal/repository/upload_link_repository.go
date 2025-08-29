package repository

import (
	"context"
	"database/sql"
	"github.com/NiClassic/go-cloud/internal/model"
	"log"
	"time"
)

type UploadLinkRepository struct {
	db *sql.DB
}

func NewLinkTokenRepository(db *sql.DB) *UploadLinkRepository {
	return &UploadLinkRepository{db: db}
}

func (l *UploadLinkRepository) Insert(ctx context.Context, hashedPassword, linkToken, name string, expiresAt time.Time) (int64, error) {
	query := `INSERT INTO upload_links (password, expires_at, link_token, name) VALUES (?, ?, ?, ?)`
	result, err := l.db.ExecContext(ctx, query, hashedPassword, expiresAt, linkToken, name)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (l *UploadLinkRepository) GetByToken(ctx context.Context, linkToken string) (*model.UploadLink, error) {
	query := `SELECT id, password, name, created_at, expires_at, link_token FROM upload_links WHERE link_token = ?`
	row := l.db.QueryRowContext(ctx, query, linkToken)

	var uploadLink model.UploadLink
	err := row.Scan(
		&uploadLink.ID,
		&uploadLink.HashedPassword,
		&uploadLink.Name,
		&uploadLink.CreatedAt,
		&uploadLink.ExpiresAt,
		&uploadLink.LinkToken,
	)
	return &uploadLink, err
}

func (l *UploadLinkRepository) GetAll(ctx context.Context) ([]*model.UploadLink, error) {
	query := `SELECT id, name, created_at, expires_at, link_token FROM upload_links`
	rows, err := l.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer func(rows *sql.Rows) {
		err = rows.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(rows)
	links := make([]*model.UploadLink, 0)

	for rows.Next() {
		var uploadLink model.UploadLink
		err = rows.Scan(
			&uploadLink.ID,
			&uploadLink.Name,
			&uploadLink.CreatedAt,
			&uploadLink.ExpiresAt,
			&uploadLink.LinkToken,
		)
		if err != nil {
			return nil, err
		}

		links = append(links, &uploadLink)
	}
	return links, nil
}
