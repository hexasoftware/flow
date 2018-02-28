package flow

import "errors"

// flow Errors
var (
	ErrNotFound  = errors.New("Entry not found")
	ErrNotAFunc  = errors.New("Is not a function")
	ErrInput     = errors.New("Invalid input")
	ErrOutput    = errors.New("Invalid output")
	ErrOperation = errors.New("Invalid operation")
)
