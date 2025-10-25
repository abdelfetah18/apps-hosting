package repositories

import (
	"context"
	"database/sql"
	"time"

	"apps-hosting.com/logging"

	"github.com/uptrace/bun"
)

type BuildStatus string

const (
	BuildStatusPending   BuildStatus = "pending"
	BuildStatusSuccessed BuildStatus = "successed"
	BuildStatusFailed    BuildStatus = "failed"
)

type Build struct {
	Id         string      `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	AppId      string      `bun:"app_id" json:"app_id"`
	Status     BuildStatus `bun:"status" json:"status"`
	ImageURL   string      `bun:"image_url" json:"image_url"`
	CommitHash string      `bun:"commit_hash" json:"commit_hash"`
	CreatedAt  time.Time   `bun:"created_at,default:now()" json:"created_at"`
}

type CreateBuildParams struct {
	Status     BuildStatus
	ImageURL   string
	CommitHash string
}

type UpdateBuildParams struct {
	Status     BuildStatus
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
	return repository.Database.NewCreateTable().Model((*Build)(nil)).IfNotExists().Exec(context.Background())
}

func (repository *BuildRepository) CreateBuild(appId string, createBuildParams CreateBuildParams) (*Build, error) {
	build := Build{
		AppId:      appId,
		Status:     createBuildParams.Status,
		ImageURL:   createBuildParams.ImageURL,
		CommitHash: createBuildParams.CommitHash,
	}
	_, err := repository.Database.
		NewInsert().
		Model(&build).
		Exec(context.Background())

	if err != nil {
		return nil, err
	}

	return &build, nil
}

func (repository *BuildRepository) UpdateBuildById(appId, buildId string, updateBuildParams UpdateBuildParams) (*Build, error) {
	build := Build{
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
		Exec(context.Background())

	if err != nil {
		return nil, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, ErrBuildNotFound
	}

	return &build, nil
}

func (repository *BuildRepository) GetBuilds(appId string) ([]Build, error) {
	builds := []Build{}
	err := repository.Database.NewSelect().
		Model(&builds).
		Where("app_id = ?", appId).
		Order("created_at DESC").
		Scan(context.Background())

	if err != nil {
		return []Build{}, err
	}

	if builds == nil {
		return []Build{}, err
	}

	return builds, nil
}

func (repository *BuildRepository) DeleteBuilds(appId string) error {
	result, err := repository.Database.
		NewDelete().
		Model(&Build{AppId: appId}).
		Where("app_id = ?", appId).
		Exec(context.Background())

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrBuildNotFound
	}

	return err
}
