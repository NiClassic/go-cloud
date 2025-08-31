package repository

import (
	"context"
	"database/sql"

	"github.com/NiClassic/go-cloud/internal/model"
)

type UserRepository struct{ baseRepo }

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{newBaseRepo(db)}
}

func (r *UserRepository) Insert(
	ctx context.Context,
	username, hashedPassword string,
) (int64, error) {
	const q = `INSERT INTO users (username, password) VALUES (?, ?)`
	res, err := r.db.ExecContext(ctx, q, username, hashedPassword)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	const q = `SELECT id, username, password FROM users WHERE username = ?`
	var u model.User
	if err := r.db.QueryRowContext(ctx, q, username).Scan(
		&u.ID, &u.Username, &u.HashedPassword,
	); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	const q = `SELECT id, username, password FROM users WHERE id = ?`
	var u model.User
	if err := r.db.QueryRowContext(ctx, q, id).Scan(
		&u.ID, &u.Username, &u.HashedPassword,
	); err != nil {
		return nil, err
	}
	return &u, nil
}
