package repositories

import (
	"context"
	"database/sql"

	"apps-hosting.com/deployservice/internal/models"
	"apps-hosting.com/logging"

	"github.com/uptrace/bun"
)

type CreateDeploymentParams struct {
	Status models.DeploymentStatus
}

type UpdateDeploymentParams struct {
	Status models.DeploymentStatus
}

type DeploymentRepository struct {
	Database *bun.DB
	Logger   logging.ServiceLogger
}

func NewDeploymentRepository(database *bun.DB, logger logging.ServiceLogger) DeploymentRepository {
	return DeploymentRepository{
		Database: database,
		Logger:   logger,
	}
}

func (repository *DeploymentRepository) CreateDeploymentsTable() (sql.Result, error) {
	repository.Logger.LogInfo("Creating deployments table.")
	return repository.Database.NewCreateTable().Model((*models.Deployment)(nil)).IfNotExists().Exec(context.Background())
}

func (repository *DeploymentRepository) CreateDeployment(ctx context.Context, buildId, appId string, createDeploymentParams CreateDeploymentParams) (*models.Deployment, error) {
	deployment := models.Deployment{
		BuildId: buildId,
		AppId:   appId,
		Status:  createDeploymentParams.Status,
	}
	_, err := repository.Database.NewInsert().Model(&deployment).Exec(ctx)
	if err != nil {
		return nil, err
	}

	return &deployment, nil
}

func (repository *DeploymentRepository) GetDeployments(ctx context.Context, appId string) ([]models.Deployment, error) {
	deployments := []models.Deployment{}
	err := repository.Database.
		NewSelect().
		Model(&deployments).
		Where("app_id = ?", appId).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return []models.Deployment{}, err
	}

	if deployments == nil {
		return []models.Deployment{}, err
	}

	return deployments, nil
}

func (repository *DeploymentRepository) DeleteDeployments(ctx context.Context, appId string) error {
	result, err := repository.Database.
		NewDelete().
		Model(&models.Deployment{AppId: appId}).
		Where("app_id = ?", appId).
		Exec(ctx)

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrDeploymentNotFound
	}

	return err
}

func (repository *DeploymentRepository) UpdateDeploymentById(ctx context.Context, deploymentId string, updateDeploymentParams UpdateDeploymentParams) (*models.Deployment, error) {
	deployment := models.Deployment{
		Status: updateDeploymentParams.Status,
	}

	result, err := repository.Database.
		NewUpdate().
		Model(&deployment).
		Column("status").
		Where("id = ?", deploymentId).
		Returning("*").
		Exec(ctx)

	if err != nil {
		return nil, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, ErrDeploymentNotFound
	}

	return &deployment, nil
}
