package repositories

import "errors"

var (
	ErrProjectNameInUse = errors.New("project with that name already exists")
	ErrProjectNotFound  = errors.New("project not found")
)
