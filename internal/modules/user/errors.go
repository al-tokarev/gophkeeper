package user

import "errors"

var (
	ErrIncorrectLogin = errors.New("incorrect login/password")
	ErrDuplicateEmail = errors.New("duplicate email")
	ErrNotFound       = errors.New("user is not found")
)
