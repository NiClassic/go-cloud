package model

type User struct {
	ID             int64  `db:"id"`
	Username       string `db:"username"`
	HashedPassword string `db:"password"`
}
