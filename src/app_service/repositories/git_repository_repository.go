package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"apps-hosting.com/logging"

	"github.com/uptrace/bun"
)

type Provider string

const (
	ProviderGithub Provider = "github"
)

type GitRepository struct {
	Id        string    `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	AppId     string    `bun:"app_id,type:uuid,notnull" json:"app_id"`
	Provider  string    `bun:"provider" json:"provider"`
	CloneURL  string    `bun:"clone_url" json:"clone_url"`
	IsPrivate bool      `bun:"is_private" json:"is_private"`
	CreatedAt time.Time `bun:"created_at,default:now()" json:"created_at"`

	App *App `bun:"rel:belongs-to,join:app_id=id"`
}

type CreateGitRepository struct {
	Provider  string
	CloneURL  string
	IsPrivate bool
}

type UpdateGitRepository struct {
	Provider  string
	CloneURL  string
	IsPrivate bool
}

type GitRepositoryRepository struct {
	Database *bun.DB
	Logger   logging.ServiceLogger
}

func NewGitRepositoryRepository(database *bun.DB, logger logging.ServiceLogger) GitRepositoryRepository {
	return GitRepositoryRepository{
		Database: database,
		Logger:   logger,
	}
}

func (repository *GitRepositoryRepository) CreateGitRepositoryRepositoryTable() (sql.Result, error) {
	repository.Logger.LogInfo("Creating git_repositories table.")
	return repository.Database.NewCreateTable().Model((*GitRepository)(nil)).IfNotExists().Exec(context.Background())
}

func (repository *GitRepositoryRepository) GetGitRepository(ctx context.Context, appId string) (*GitRepository, error) {
	gitRepository := GitRepository{}
	err := repository.Database.
		NewSelect().
		Model(&gitRepository).
		Where("app_id = ?", appId).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGitRepositoryNotFound
		}
		return nil, err
	}

	return &gitRepository, nil
}

func (repository *GitRepositoryRepository) CreateGitRepository(ctx context.Context, appId string, createGitRepository CreateGitRepository) (*GitRepository, error) {
	gitRepository := GitRepository{
		AppId:     appId,
		Provider:  createGitRepository.Provider,
		CloneURL:  createGitRepository.CloneURL,
		IsPrivate: createGitRepository.IsPrivate,
	}

	_, err := repository.Database.
		NewInsert().
		Model(&gitRepository).
		Exec(ctx)

	if err != nil {
		return nil, err
	}

	return &gitRepository, nil
}

func (repository *GitRepositoryRepository) UpdateGitRepository(ctx context.Context, appId string, updateGitRepository UpdateGitRepository) (*GitRepository, error) {
	gitRepository := GitRepository{
		Provider:  updateGitRepository.Provider,
		CloneURL:  updateGitRepository.CloneURL,
		IsPrivate: updateGitRepository.IsPrivate,
	}

	result, err := repository.Database.
		NewUpdate().
		Model(&gitRepository).
		Column("provider", "clone_url", "is_private").
		Where("app_id = ?", appId).
		Returning("*").
		Exec(ctx)

	if err != nil {
		return nil, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, ErrGitRepositoryNotFound
	}

	return &gitRepository, nil
}
