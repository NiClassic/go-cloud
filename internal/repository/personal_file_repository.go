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

func (p *PersonalFileRepository) Insert(ctx context.Context, name, mimeType, location, hash string, userId, size int64, folderID int64) (int64, error) {
	const q = `INSERT INTO files (user_id, name, size, mime_type, location, hash, folder_id) VALUES (?, ?, ?, ?, ?, ?, ?)`
	res, err := p.db.ExecContext(ctx, q, userId, name, size, mimeType, location, hash, folderID)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (p *PersonalFileRepository) GetByUser(ctx context.Context, id int64) ([]*model.File, error) {
	const q = `SELECT id, user_id, name, size, mime_type, created_at, location, hash, folder_id FROM files WHERE user_id = ?`
	rows, err := p.db.QueryContext(ctx, q, id)
	if err != nil {
		return nil, err
	}
	defer p.closeRows(rows)

	var files []*model.File
	for rows.Next() {
		var f model.File
		if err := rows.Scan(
			&f.ID, &f.UserID, &f.Name, &f.Size, &f.MimeType, &f.CreatedAt, &f.Location, &f.Hash, &f.FolderID); err != nil {
			return nil, err
		}
		files = append(files, &f)
	}
	return files, rows.Err()
}

func (p *PersonalFileRepository) GetByUserAndFolder(ctx context.Context, userID int64, folderID int64) ([]*model.File, error) {
	const q = `SELECT id, user_id, name, size, mime_type, created_at, location, hash, folder_id FROM files WHERE user_id = ? AND folder_id = ?`

	rows, err := p.db.QueryContext(ctx, q, userID, folderID)
	if err != nil {
		return nil, err
	}
	defer p.closeRows(rows)

	var files []*model.File
	for rows.Next() {
		var f model.File
		if err := rows.Scan(
			&f.ID, &f.UserID, &f.Name, &f.Size, &f.MimeType, &f.CreatedAt, &f.Location, &f.Hash, &f.FolderID); err != nil {
			return nil, err
		}
		files = append(files, &f)
	}
	return files, rows.Err()
}

func (p *PersonalFileRepository) GetById(ctx context.Context, id int64) (*model.File, error) {
	const q = `SELECT id, user_id, name, size, mime_type, created_at, location, hash, folder_id FROM files WHERE id = ?`
	var f model.File
	if err := p.db.QueryRowContext(ctx, q, id).Scan(&f.ID, &f.UserID, &f.Name, &f.Size, &f.MimeType, &f.CreatedAt, &f.Location, &f.Hash, &f.FolderID); err != nil {
		return nil, err
	}
	return &f, nil
}

func (p *PersonalFileRepository) UpdateFolder(ctx context.Context, fileID int64, folderID int64) error {
	const q = `UPDATE files SET folder_id = ? WHERE id = ?`
	_, err := p.db.ExecContext(ctx, q, folderID, fileID)
	return err
}

func (p *PersonalFileRepository) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM files WHERE id = ?`
	_, err := p.db.ExecContext(ctx, q, id)
	return err
}
