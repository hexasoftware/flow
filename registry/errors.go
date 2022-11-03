package registry

import "errors"

// flow Errors
var (
	ErrNotFound = errors.New("Entry not found")
	ErrNotAFunc = errors.New("Is not a function")
	ErrOutput   = errors.New("Invalid output")
)
