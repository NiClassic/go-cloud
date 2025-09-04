package repository

import (
	"context"
	"database/sql"
	"github.com/NiClassic/go-cloud/internal/model"
)

type PersonalFileRepository struct{ baseRepo }

func NewPersonalFileRepository(db *sql.DB) *PersonalFileRepository {
	return &PersonalFileRepository{newBaseRepo(db)}
}

func (p *PersonalFileRepository) Insert(ctx context.Context, name, mimeType, location, hash string, userId, size int64) (int64, error) {
	const q = `INSERT INTO files (user_id, name, size, mime_type, location, hash) VALUES (?, ?, ?, ?, ?, ?)`
	res, err := p.db.ExecContext(ctx, q, userId, name, size, mimeType, location, hash)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (p *PersonalFileRepository) GetByUser(ctx context.Context, id int64) ([]*model.File, error) {
	const q = `SELECT * FROM files WHERE user_id = ?`
	rows, err := p.db.QueryContext(ctx, q, id)
	if err != nil {
		return nil, err
	}
	defer p.closeRows(rows)

	var files []*model.File
	for rows.Next() {
		var f model.File
		if err := rows.Scan(
			&f.ID, &f.UserID, &f.Name, &f.Size, &f.MimeType, &f.CreatedAt, &f.Location, &f.Hash); err != nil {
			return nil, err
		}
		files = append(files, &f)
	}
	return files, rows.Err()
}
