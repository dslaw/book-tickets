package cache

import "errors"

var (
	ErrAlreadyHasHold = errors.New("The ticket is already held")
	ErrNotFound       = errors.New("The key or field was not found")
)
