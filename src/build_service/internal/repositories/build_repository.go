package repositories

import (
	"context"
	"database/sql"

	"apps-hosting.com/buildservice/internal/models"

	"apps-hosting.com/logging"

	"github.com/uptrace/bun"
)

type CreateBuildParams struct {
	Status     models.BuildStatus
	ImageURL   string
	CommitHash string
}

type UpdateBuildParams struct {
	Status     models.BuildStatus
	ImageURL   string
	CommitHash string
}

type BuildRepository struct {
	Database *bun.DB
	Logger   logging.ServiceLogger
}

func NewBuildRepository(database *bun.DB, logger logging.ServiceLogger) BuildRepository {
	return BuildRepository{
		Database: database,
		Logger:   logger,
	}
}

func (repository *BuildRepository) CreateBuildsTable() (sql.Result, error) {
	repository.Logger.LogInfo("Creating builds table.")
	return repository.Database.NewCreateTable().Model((*models.Build)(nil)).IfNotExists().Exec(context.Background())
}

func (repository *BuildRepository) CreateBuild(ctx context.Context, appId string, createBuildParams CreateBuildParams) (*models.Build, error) {
	build := models.Build{
		AppId:      appId,
		Status:     createBuildParams.Status,
		ImageURL:   createBuildParams.ImageURL,
		CommitHash: createBuildParams.CommitHash,
	}
	_, err := repository.Database.
		NewInsert().
		Model(&build).
		Exec(ctx)

	if err != nil {
		return nil, err
	}

	return &build, nil
}

func (repository *BuildRepository) UpdateBuildById(ctx context.Context, appId, buildId string, updateBuildParams UpdateBuildParams) (*models.Build, error) {
	build := models.Build{
		Status:     updateBuildParams.Status,
		ImageURL:   updateBuildParams.ImageURL,
		CommitHash: updateBuildParams.CommitHash,
	}

	result, err := repository.Database.
		NewUpdate().
		Model(&build).
		Column("status", "image_url", "commit_hash").
		Where("id = ? and app_id = ?", buildId, appId).
		Returning("*").
		Exec(ctx)

	if err != nil {
		return nil, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, ErrBuildNotFound
	}

	return &build, nil
}

func (repository *BuildRepository) GetBuilds(ctx context.Context, appId string) ([]models.Build, error) {
	builds := []models.Build{}
	err := repository.Database.NewSelect().
		Model(&builds).
		Where("app_id = ?", appId).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return []models.Build{}, err
	}

	if builds == nil {
		return []models.Build{}, err
	}

	return builds, nil
}

func (repository *BuildRepository) DeleteBuilds(ctx context.Context, appId string) error {
	result, err := repository.Database.
		NewDelete().
		Model(&models.Build{AppId: appId}).
		Where("app_id = ?", appId).
		Exec(ctx)

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrBuildNotFound
	}

	return err
}
