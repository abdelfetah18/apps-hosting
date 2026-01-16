package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"apps-hosting.com/logging"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/uptrace/bun"
)

type AppStatus string

const (
	AppStatusBuilding     AppStatus = "building"
	AppStatusDeploying    AppStatus = "deploying"
	AppStatusDeployed     AppStatus = "deployed"
	AppStatusBuildFailed  AppStatus = "build_failed"
	AppStatusDeployFailed AppStatus = "deploy_failed"
)

type App struct {
	Id         string    `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	ProjectId  string    `bun:"project_id" json:"project_id"`
	Name       string    `bun:"name,unique" json:"name"`
	DomainName string    `bun:"domain_name,unique" json:"domain_name"`
	Runtime    string    `bun:"runtime" json:"runtime"`
	RepoURL    string    `bun:"repo_url" json:"repo_url"`
	BuildCMD   string    `bun:"build_cmd" json:"build_cmd"`
	StartCMD   string    `bun:"start_cmd" json:"start_cmd"`
	CreatedAt  time.Time `bun:"created_at,default:now()" json:"created_at"`
}

type CreateAppParams struct {
	Name       string
	Runtime    string
	RepoURL    string
	StartCMD   string
	BuildCMD   string
	DomainName string
}

type UpdateAppParams struct {
	Name     string
	StartCMD string
	BuildCMD string
}

var Runtimes = []string{"NodeJS"}

type AppRepository struct {
	Database *bun.DB
	Logger   logging.ServiceLogger
}

func NewAppRepository(database *bun.DB, logger logging.ServiceLogger) AppRepository {
	return AppRepository{
		Database: database,
		Logger:   logger,
	}
}

func (repository *AppRepository) CreateAppsTable() (sql.Result, error) {
	repository.Logger.LogInfo("Creating apps table.")
	return repository.Database.NewCreateTable().Model((*App)(nil)).IfNotExists().Exec(context.Background())
}

func (repository *AppRepository) CreateApp(ctx context.Context, projectId string, createAppParams CreateAppParams) (*App, error) {
	app := App{
		Name:       createAppParams.Name,
		Runtime:    createAppParams.Runtime,
		ProjectId:  projectId,
		RepoURL:    createAppParams.RepoURL,
		StartCMD:   createAppParams.StartCMD,
		BuildCMD:   createAppParams.BuildCMD,
		DomainName: createAppParams.DomainName,
	}
	_, err := repository.Database.NewInsert().Model(&app).Exec(ctx)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				if pgErr.ConstraintName == "apps_name_key" {
					return nil, ErrAppNameInUse
				}
				if pgErr.ConstraintName == "apps_domain_name_key" {
					return nil, ErrDomainNameInUse
				}
			}
		}

		return nil, err
	}

	return &app, nil
}

func (repository *AppRepository) UpdateApp(ctx context.Context, projectId, appId string, updateAppParams UpdateAppParams) (*App, error) {
	app := App{
		Name:     updateAppParams.Name,
		StartCMD: updateAppParams.StartCMD,
		BuildCMD: updateAppParams.BuildCMD,
	}

	result, err := repository.Database.
		NewUpdate().
		Model(&app).
		Column("name", "build_cmd", "start_cmd").
		Where("id = ? and project_id = ?", appId, projectId).
		Returning("*").
		Exec(ctx)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				if pgErr.ConstraintName == "apps_name_key" {
					return nil, ErrAppNameInUse
				}
				if pgErr.ConstraintName == "apps_domain_name_key" {
					return nil, ErrDomainNameInUse
				}
			}
		}

		return nil, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, ErrAppNotFound
	}

	return &app, nil
}

func (repository *AppRepository) GetAppById(ctx context.Context, projectId, appId string) (*App, error) {
	app := App{}
	err := repository.Database.
		NewSelect().
		Model(&app).
		Where("id = ? and project_id = ?", appId, projectId).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAppNotFound
		}
		return nil, err
	}

	return &app, nil
}

func (repository *AppRepository) GetApps(ctx context.Context, projectId string) ([]App, error) {
	var apps []App
	err := repository.Database.
		NewSelect().
		Model(&apps).
		Where("project_id = ?", projectId).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return []App{}, err
	}

	if apps == nil {
		return []App{}, err
	}

	return apps, nil
}

func (repository *AppRepository) DeleteAppById(ctx context.Context, projectId string, appId string) error {
	result, err := repository.Database.
		NewDelete().
		Model(&App{Id: appId, ProjectId: projectId}).
		Where("id = ? AND project_id = ?", appId, projectId).
		Exec(ctx)

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrAppNotFound
	}

	return err
}

func (repository *AppRepository) DeleteAppsByProjectId(ctx context.Context, projectId string) error {
	_, err := repository.Database.
		NewDelete().
		Model(&App{ProjectId: projectId}).
		Where("project_id = ?", projectId).
		Exec(ctx)
	return err
}

func (repository *AppRepository) GetProjectsAppsCounts(
	ctx context.Context,
	projectIds []string,
) (map[string]int32, error) {

	if len(projectIds) == 0 {
		return map[string]int32{}, nil
	}

	type result struct {
		ProjectId string `bun:"project_id"`
		Count     int32  `bun:"apps_count"`
	}

	var rows []result

	err := repository.Database.
		NewSelect().
		Model((*App)(nil)).
		Column("project_id").
		ColumnExpr("COUNT(*) AS apps_count").
		Where("project_id IN (?)", bun.In(projectIds)).
		Group("project_id").
		Scan(ctx, &rows)

	if err != nil {
		return nil, err
	}

	projectAppsCount := make(map[string]int32, len(rows))
	for _, row := range rows {
		projectAppsCount[row.ProjectId] = row.Count
	}

	return projectAppsCount, nil
}
