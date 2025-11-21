package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"apps-hosting.com/logging"

	"github.com/uptrace/bun"
)

type EnvironmentVariable struct {
	Id        string    `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	AppId     string    `bun:"app_id,type:uuid,notnull" json:"app_id"`
	Value     string    `bun:"value" json:"value"` // Store json object of env vars
	CreatedAt time.Time `bun:"created_at,default:now()" json:"created_at"`

	App *App `bun:"rel:belongs-to,join:app_id=id"`
}

type CreateEnvironmentVariableParams struct {
	Value string
}

type UpdateEnvironmentVariableParams struct {
	Value string
}

type EnvironmentVariablesRepository struct {
	Database *bun.DB
	Logger   logging.ServiceLogger
}

func NewEnvironmentVariablesRepository(database *bun.DB, logger logging.ServiceLogger) EnvironmentVariablesRepository {
	return EnvironmentVariablesRepository{
		Database: database,
		Logger:   logger,
	}
}

func (repository *EnvironmentVariablesRepository) CreateEnvironmentVariablesTable() (sql.Result, error) {
	repository.Logger.LogInfo("Creating environment_variables table.")
	return repository.Database.NewCreateTable().Model((*EnvironmentVariable)(nil)).IfNotExists().Exec(context.Background())
}

func (repository *EnvironmentVariablesRepository) GetEnvironmentVariable(ctx context.Context, appId string) (*EnvironmentVariable, error) {
	environmentVariable := EnvironmentVariable{}
	err := repository.Database.
		NewSelect().
		Model(&environmentVariable).
		Where("app_id = ?", appId).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEnvVarNotFound
		}
		return nil, err
	}

	return &environmentVariable, nil
}

func (repository *EnvironmentVariablesRepository) CreateEnvironmentVariables(ctx context.Context, appId string, createEnvironmentVariableParams CreateEnvironmentVariableParams) (*EnvironmentVariable, error) {
	environmentVariable := EnvironmentVariable{
		AppId: appId,
		Value: createEnvironmentVariableParams.Value,
	}

	_, err := repository.Database.
		NewInsert().
		Model(&environmentVariable).
		Exec(ctx)

	if err != nil {
		return nil, err
	}

	return &environmentVariable, nil
}

func (repository *EnvironmentVariablesRepository) UpdateEnvironmentVariables(ctx context.Context, appId string, updateEnvironmentVariableParams UpdateEnvironmentVariableParams) (*EnvironmentVariable, error) {
	environmentVariable := EnvironmentVariable{
		Value: updateEnvironmentVariableParams.Value,
	}

	result, err := repository.Database.
		NewUpdate().
		Model(&environmentVariable).
		Column("value").
		Where("app_id = ?", appId).
		Returning("*").
		Exec(ctx)

	if err != nil {
		return nil, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, ErrEnvVarNotFound
	}

	return &environmentVariable, nil
}

func (repository *EnvironmentVariablesRepository) DeleteEnvironmentVariableByAppId(ctx context.Context, appId string) error {
	_, err := repository.Database.
		NewDelete().
		Model(&EnvironmentVariable{AppId: appId}).
		Where("app_id = ?", appId).
		Exec(ctx)
	return err
}

func (repository *EnvironmentVariablesRepository) DeleteEnvironmentVariablesByAppIds(ctx context.Context, appIds []string) error {
	_, err := repository.Database.
		NewDelete().
		Model((*EnvironmentVariable)(nil)).
		Where("app_id IN (?)", bun.In(appIds)).
		Exec(ctx)
	return err
}
