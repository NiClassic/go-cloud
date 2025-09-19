package service

import "errors"

var (
	ErrEmptyCredentials   = errors.New("username and password required")
	ErrEmptyLinkFields    = errors.New("name and password required")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrSessionInvalid     = errors.New("session invalid")
	ErrLinkNotFound       = errors.New("upload link not found")
	ErrLinkExpired        = errors.New("upload link expired")
	ErrInvalidPassword    = errors.New("invalid password")
)
