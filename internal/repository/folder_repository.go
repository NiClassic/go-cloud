package repository

import (
	"context"
	"database/sql"

	"github.com/NiClassic/go-cloud/internal/model"
)

type FolderRepository struct{ baseRepo }

func NewFolderRepository(db *sql.DB) *FolderRepository {
	return &FolderRepository{newBaseRepo(db)}
}

func (r *FolderRepository) Insert(ctx context.Context, userID int64, parentID *int64, name string, path string) (int64, error) {
	const q = `INSERT INTO folders (user_id, parent_id, name, path) VALUES (?, ?, ?, ?)`

	var res sql.Result
	var err error

	if parentID == nil {
		res, err = r.db.ExecContext(ctx, q, userID, nil, name, path)
	} else {
		res, err = r.db.ExecContext(ctx, q, userID, *parentID, name, path)
	}

	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *FolderRepository) GetByID(ctx context.Context, id int64) (*model.Folder, error) {
	const q = `SELECT id, user_id, parent_id, name, path, created_at, updated_at FROM folders WHERE id = ?`
	var f model.Folder
	if err := r.db.QueryRowContext(ctx, q, id).Scan(
		&f.ID, &f.UserID, &f.ParentID, &f.Name, &f.Path, &f.CreatedAt, &f.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FolderRepository) GetByPath(ctx context.Context, path string) (*model.Folder, error) {
	const q = `SELECT id, user_id, parent_id, name, path, created_at, updated_at FROM folders WHERE path = ?`
	var f model.Folder
	if err := r.db.QueryRowContext(ctx, q, path).Scan(
		&f.ID, &f.UserID, &f.ParentID, &f.Name, &f.Path, &f.CreatedAt, &f.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FolderRepository) GetByUser(ctx context.Context, userID int64) ([]*model.Folder, error) {
	const q = `SELECT id, user_id, parent_id, name, path, created_at, updated_at
	FROM folders WHERE user_id = ? AND parent_id IS NULL ORDER BY name`

	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer r.closeRows(rows)

	var folders []*model.Folder
	for rows.Next() {
		var f model.Folder
		if err := rows.Scan(&f.ID, &f.UserID, &f.ParentID, &f.Name, &f.Path, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		folders = append(folders, &f)
	}
	return folders, rows.Err()
}

func (r *FolderRepository) GetByUserAndParent(ctx context.Context, userID int64, parentID int64) ([]*model.Folder, error) {
	const q = `SELECT id, user_id, parent_id, name, path, created_at, updated_at 
		     FROM folders WHERE user_id = ? AND parent_id = ? ORDER BY name`

	rows, err := r.db.QueryContext(ctx, q, userID, parentID)
	if err != nil {
		return nil, err
	}
	defer r.closeRows(rows)

	var folders []*model.Folder
	for rows.Next() {
		var f model.Folder
		if err := rows.Scan(&f.ID, &f.UserID, &f.ParentID, &f.Name, &f.Path, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		folders = append(folders, &f)
	}
	return folders, rows.Err()
}

func (r *FolderRepository) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM folders WHERE id = ?`
	_, err := r.db.ExecContext(ctx, q, id)
	return err
}

func (r *FolderRepository) UpdatePath(ctx context.Context, folderID int64, path string) error {
	const q = `UPDATE folders SET path = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := r.db.ExecContext(ctx, q, path, folderID)
	return err
}

func (r *FolderRepository) UpdateParent(ctx context.Context, folderID int64, newParentID int64) error {
	const q = `UPDATE folders SET parent_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := r.db.ExecContext(ctx, q, newParentID, folderID)
	return err
}
