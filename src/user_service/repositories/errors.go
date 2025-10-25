package repositories

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrUsernameInUse       = errors.New("user with that username already exists")
	ErrEmailInUse          = errors.New("user with that email already exists")
	ErrUserSessionNotFound = errors.New("user session not found")
)
