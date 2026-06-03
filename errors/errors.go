package errors

import "errors"

var (
	ErrNotFound = errors.New("resource not found")
	ErrInvalid  = errors.New("invalid request")
)
