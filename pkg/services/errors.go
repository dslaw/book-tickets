package services

import "errors"

var (
	ErrInvalidHoldID  = errors.New("Invalid hold id")
	ErrHoldIDMismatch = errors.New("The given hold id does not match")
)
