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

type ProjectRepositoryInterface interface {
	CreateProjectsTable() error
	CreateProject(userId string, createProjectParams CreateProjectParams) (*Project, error)
	GetProjectById(userId, projectId string) (*Project, error)
	GetProjects(userId string) ([]Project, error)
	DeleteProjectById(userId, projectId string) error
}

type Project struct {
	Id        string    `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	Name      string    `bun:"name,unique" json:"name"`
	UserId    string    `bun:"user_id" json:"user_id"`
	CreatedAt time.Time `bun:"created_at,default:now()" json:"created_at"`
}

type CreateProjectParams struct {
	Name string
}

type ProjectRepository struct {
	Database *bun.DB
	Logger   logging.ServiceLogger
}

func NewProjectRepository(database *bun.DB, logger logging.ServiceLogger) ProjectRepository {
	return ProjectRepository{
		Database: database,
		Logger:   logger,
	}
}

func (repository *ProjectRepository) CreateProjectsTable() error {
	repository.Logger.LogInfo("Creating projects table.")
	_, err := repository.Database.NewCreateTable().Model((*Project)(nil)).IfNotExists().Exec(context.Background())
	return err
}

func (repository *ProjectRepository) CreateProject(userId string, createProjectParams CreateProjectParams) (*Project, error) {
	project := Project{
		UserId: userId,
		Name:   createProjectParams.Name,
	}

	_, err := repository.Database.NewInsert().Model(&project).Exec(context.Background())
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				if pgErr.ConstraintName == "projects_name_key" {
					return nil, ErrProjectNameInUse
				}
			}
		}

		return nil, err
	}

	return &project, nil
}

func (repository *ProjectRepository) GetProjectById(userId, projectId string) (*Project, error) {
	project := Project{}

	err := repository.Database.
		NewSelect().
		Model(&project).
		Where("id = ? and user_id = ?", projectId, userId).
		Scan(context.Background())

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	return &project, nil
}

func (repository *ProjectRepository) GetProjects(userId string) ([]Project, error) {
	var projects []Project
	err := repository.Database.NewSelect().
		Model(&projects).
		Where("user_id = ?", userId).
		Order("created_at DESC").
		Scan(context.Background())
	if err != nil {
		return []Project{}, err
	}

	if projects == nil {
		return []Project{}, err
	}

	return projects, nil
}

func (repository *ProjectRepository) DeleteProjectById(userId, projectId string) error {
	result, err := repository.Database.
		NewDelete().
		Model(&Project{}).
		Where("id = ? AND user_id = ?", projectId, userId).
		Exec(context.Background())

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrProjectNotFound
	}

	return err
}
