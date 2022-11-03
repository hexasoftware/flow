package flow

import "errors"

// flow Errors
var (
	ErrNotFound  = errors.New("entry not found")
	ErrNotAFunc  = errors.New("is not a function")
	ErrInput     = errors.New("invalid input")
	ErrOutput    = errors.New("invalid output")
	ErrOperation = errors.New("invalid operation")
)
