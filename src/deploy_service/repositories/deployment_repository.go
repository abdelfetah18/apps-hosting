package repositories

import (
	"context"
	"database/sql"
	"time"

	"apps-hosting.com/logging"

	"github.com/uptrace/bun"
)

type DeploymentStatus string

const (
	DeploymentStatusPending   DeploymentStatus = "pending"
	DeploymentStatusSuccessed DeploymentStatus = "successed"
	DeploymentStatusFailed    DeploymentStatus = "failed"
)

type Deployment struct {
	Id        string           `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	BuildId   string           `bun:"build_id" json:"build_id"`
	AppId     string           `bun:"app_id" json:"app_id"`
	Status    DeploymentStatus `bun:"status" json:"status"`
	CreatedAt time.Time        `bun:"created_at,default:now()" json:"created_at"`
}

type CreateDeploymentParams struct {
	Status DeploymentStatus
}

type UpdateDeploymentParams struct {
	Status DeploymentStatus
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
	return repository.Database.NewCreateTable().Model((*Deployment)(nil)).IfNotExists().Exec(context.Background())
}

func (repository *DeploymentRepository) CreateDeployment(ctx context.Context, buildId, appId string, createDeploymentParams CreateDeploymentParams) (*Deployment, error) {
	deployment := Deployment{
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

func (repository *DeploymentRepository) GetDeployments(ctx context.Context, appId string) ([]Deployment, error) {
	deployments := []Deployment{}
	err := repository.Database.
		NewSelect().
		Model(&deployments).
		Where("app_id = ?", appId).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return []Deployment{}, err
	}

	if deployments == nil {
		return []Deployment{}, err
	}

	return deployments, nil
}

func (repository *DeploymentRepository) DeleteDeployments(ctx context.Context, appId string) error {
	result, err := repository.Database.
		NewDelete().
		Model(&Deployment{AppId: appId}).
		Where("app_id = ?", appId).
		Exec(ctx)

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrDeploymentNotFound
	}

	return err
}

func (repository *DeploymentRepository) UpdateDeploymentById(ctx context.Context, deploymentId string, updateDeploymentParams UpdateDeploymentParams) (*Deployment, error) {
	deployment := Deployment{
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
