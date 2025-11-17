package repositories

import "errors"

var (
	ErrEnvVarNotFound        = errors.New("environment variable not found")
	ErrAppNameInUse          = errors.New("app with that name already exists")
	ErrDomainNameInUse       = errors.New("domain with that name already exists")
	ErrAppNotFound           = errors.New("app not found")
	ErrGitRepositoryNotFound = errors.New("git repository not found")
)
