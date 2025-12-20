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

type User struct {
	Id                          string    `bun:"id,pk,default:gen_random_uuid()" json:"id"`
	Username                    string    `bun:"username,unique" json:"username"`
	Password                    string    `bun:"password" json:"-"`
	Email                       string    `bun:"email,unique" json:"email"`
	GithubAppInstalled          bool      `bun:"github_app_installed" json:"github_app_installed"`
	GithubAccessToken           string    `bun:"github_access_token" json:"github_access_token"`
	GithhubAccessTokenExpiresAt time.Time `bun:"github_access_token_expires_at,default:null" json:"github_access_token_expires_at"`
	GithubRefreshToken          string    `bun:"github_refresh_token" json:"github_refresh_token"`
	GithubRefreshTokenExpiresAt time.Time `bun:"github_refresh_token_expires_at,default:null" json:"github_refresh_token_expires_at"`
	CreatedAt                   time.Time `bun:"created_at,default:now()" json:"created_at"`
}

type CreateUserParams struct {
	Username string
	Password string
	Email    string
}

type UpdateUserGithubParams struct {
	GithubAppInstalled          bool
	GithubAccessToken           string
	GithhubAccessTokenExpiresAt time.Time
	GithubRefreshToken          string
	GithubRefreshTokenExpiresAt time.Time
}

type UserRepository struct {
	Database *bun.DB
	Logger   logging.ServiceLogger
}

func NewUserRepository(database *bun.DB, logger logging.ServiceLogger) UserRepository {
	return UserRepository{
		Database: database,
		Logger:   logger,
	}
}

func (repository *UserRepository) CreateUsersTable() (sql.Result, error) {
	repository.Logger.LogInfo("Creating users table.")
	return repository.Database.NewCreateTable().Model((*User)(nil)).IfNotExists().Exec(context.Background())
}

func (repository *UserRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user := User{}

	err := repository.Database.
		NewSelect().
		Model(&user).
		Where("email = ?", email).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return &user, nil
}

func (repository *UserRepository) CreateUser(ctx context.Context, createUserParams CreateUserParams) (*User, error) {
	user := User{
		Username: createUserParams.Username,
		Email:    createUserParams.Email,
		Password: createUserParams.Password,
	}

	_, err := repository.Database.NewInsert().Model(&user).Exec(ctx)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				if pgErr.ConstraintName == "users_username_key" {
					return nil, ErrUsernameInUse
				}
				if pgErr.ConstraintName == "users_email_key" {
					return nil, ErrEmailInUse
				}
			}
		}

		return nil, err
	}

	return &user, nil
}

func (repository *UserRepository) GetUserById(ctx context.Context, id string) (*User, error) {
	user := User{}

	err := repository.Database.
		NewSelect().
		Model(&user).
		Where("id = ?", id).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return &user, nil
}

func (repository *UserRepository) UpdateUserGithub(ctx context.Context, userId string, updateUserGithubParams UpdateUserGithubParams) (*User, error) {
	user := User{
		GithubAppInstalled:          updateUserGithubParams.GithubAppInstalled,
		GithubAccessToken:           updateUserGithubParams.GithubAccessToken,
		GithubRefreshToken:          updateUserGithubParams.GithubRefreshToken,
		GithhubAccessTokenExpiresAt: updateUserGithubParams.GithhubAccessTokenExpiresAt,
		GithubRefreshTokenExpiresAt: updateUserGithubParams.GithubRefreshTokenExpiresAt,
	}

	result, err := repository.Database.
		NewUpdate().
		Model(&user).
		Column("github_app_installed", "github_access_token", "github_refresh_token", "github_access_token_expires_at", "github_refresh_token_expires_at").
		Where("id = ?", userId).
		Returning("*").
		Exec(ctx)

	if err != nil {
		return nil, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, ErrUserNotFound
	}

	return &user, nil
}
