package repository

import "database/sql"

type baseRepo struct{ db *sql.DB }

func newBaseRepo(db *sql.DB) baseRepo { return baseRepo{db: db} }

func (b *baseRepo) closeRows(rows *sql.Rows) { _ = rows.Close() }
