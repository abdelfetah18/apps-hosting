package gitclient

import (
	"fmt"
	"os"

	"apps-hosting.com/logging"

	"github.com/go-git/go-git/v5"
	"github.com/google/uuid"
)

type GitRepo struct {
	Id             string
	Path           string
	LastCommitHash string
}

type GitRepoManager struct {
}

func NewGitRepoManager() GitRepoManager {
	return GitRepoManager{}
}

func (gitRepoManager *GitRepoManager) Clone(repoURL string, userAppLogger logging.UserAppLogger) (*GitRepo, error) {
	repoId := uuid.New().String()
	localPath := fmt.Sprintf("/shared/repos/%s", repoId)

	if err := os.MkdirAll(localPath, os.ModePerm); err != nil {
		return nil, err
	}

	userAppLogger.LogInfo(fmt.Sprintf("Cloning %s into %s...", repoURL, localPath))
	repo, err := git.PlainClone(localPath, false, &git.CloneOptions{
		URL:      repoURL,
		Progress: userAppLogger,
	})

	if err != nil {
		userAppLogger.LogError(err.Error())
		return nil, err
	}

	// Get the HEAD reference
	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	// Get the commit object
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit object: %w", err)
	}

	userAppLogger.LogInfo("Clone successful!")

	return &GitRepo{
		Id:             repoId,
		Path:           localPath,
		LastCommitHash: commit.Hash.String(),
	}, nil
}
