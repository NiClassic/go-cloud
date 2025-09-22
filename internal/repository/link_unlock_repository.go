package repository

import (
	"context"
	"database/sql"
	"github.com/NiClassic/go-cloud/internal/model"
	"time"
)

type LinkUnlockRepository struct{ baseRepo }

func NewLinkUnlockRepository(db *sql.DB) *LinkUnlockRepository {
	return &LinkUnlockRepository{newBaseRepo(db)}
}

func (r *LinkUnlockRepository) Insert(ctx context.Context, userID, uploadLinkID int64, expiry time.Time) (int64, error) {
	const q = `INSERT INTO link_unlocks (user_id, upload_link_id, expiry) VALUES (?, ?, ?)`
	res, err := r.db.ExecContext(ctx, q, userID, uploadLinkID, expiry)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *LinkUnlockRepository) GetByUser(ctx context.Context, userID int64) ([]*model.LinkUnlock, error) {
	const q = `SELECT * FROM link_unlocks WHERE user_id = ?`
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer r.closeRows(rows)

	var unlocks []*model.LinkUnlock
	for rows.Next() {
		var u model.LinkUnlock
		if err := rows.Scan(&u.ID, &u.UserID, &u.UploadLinkID, &u.CreatedAt, &u.Valid, &u.Expiry); err != nil {
			return nil, err
		}
		unlocks = append(unlocks, &u)
	}
	return unlocks, rows.Err()
}

func (r *LinkUnlockRepository) GetByUserAndUploadLinkID(ctx context.Context, userID, uploadLinkID int64) (*model.LinkUnlock, error) {
	const q = `SELECT * FROM link_unlocks WHERE user_id = ? AND upload_link_id = ?`
	var u model.LinkUnlock
	if err := r.db.QueryRowContext(ctx, q, userID, uploadLinkID).Scan(&u.ID, &u.UserID, &u.UploadLinkID, &u.CreatedAt, &u.Valid, &u.Expiry); err != nil {
		return nil, err
	}
	return &u, nil
}
