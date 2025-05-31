package db

import "errors"

var (
	ErrMaxRetriesReached = errors.New("error reached max retries")
)
