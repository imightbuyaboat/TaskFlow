package db

import "errors"

var (
	ErrUserAlreadyExist  = errors.New("user already exist")
	ErrIncorrectPassword = errors.New("incorrect password")
	ErrNoRows            = errors.New("no rows selected")
)
