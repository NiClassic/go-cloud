package repository

import (
	"context"
	"database/sql"
	"github.com/NiClassic/go-cloud/internal/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (u *UserRepository) Insert(ctx context.Context, username, hashedPassword string) (int64, error) {
	query := `INSERT INTO users (username, password) VALUES (?, ?)`
	result, err := u.db.ExecContext(ctx, query, username, hashedPassword)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
func (u *UserRepository) GetByUsername(_ context.Context, username string) (*model.User, error) {
	query := `SELECT * FROM users WHERE username = ?`
	row := u.db.QueryRow(query, username)
	var user model.User
	err := row.Scan(&user.ID, &user.Username, &user.HashedPassword)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (u *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	query := `SELECT * FROM users WHERE id = ?`
	row := u.db.QueryRow(query, id)

	var user model.User
	err := row.Scan(&user.ID, &user.Username, &user.HashedPassword)
	return &user, err
}
