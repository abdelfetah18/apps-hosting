package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"apps-hosting.com/logging"

	"github.com/uptrace/bun"
)

type UserSession struct {
	Id          string    `bun:"id,pk,default:gen_random_uuid()" json:"id"`
	AccessToken string    `bun:"access_token" json:"access_token"`
	UserId      string    `bun:"user_id,type:uuid,notnull" json:"user_id"`
	CreatedAt   time.Time `bun:"created_at,default:now()" json:"created_at"`

	User *User `bun:"rel:belongs-to,join:user_id=id"`
}

type CreateUserSessionParams struct {
	AccessToken string
}

type UserSessionRepository struct {
	Database *bun.DB
	Logger   logging.ServiceLogger
}

func NewUserSessionRepository(database *bun.DB, logger logging.ServiceLogger) UserSessionRepository {
	return UserSessionRepository{Database: database, Logger: logger}
}

func (userSessionRepository *UserSessionRepository) CreateUserSessionsTable() (sql.Result, error) {
	userSessionRepository.Logger.LogInfo("Creating user_sessions table.")
	return userSessionRepository.Database.
		NewCreateTable().
		Model((*UserSession)(nil)).
		IfNotExists().
		Exec(context.Background())
}

func (userSessionRepository *UserSessionRepository) CreateUserSession(ctx context.Context, userId string, createUserSessionParams CreateUserSessionParams) (*UserSession, error) {
	userSession := UserSession{
		AccessToken: createUserSessionParams.AccessToken,
		UserId:      userId,
	}

	_, err := userSessionRepository.Database.
		NewInsert().
		Model(&userSession).
		Exec(ctx)

	if err != nil {
		return nil, err
	}

	return &userSession, nil
}

func (userSessionRepository *UserSessionRepository) GetUserSessionByAccessToken(ctx context.Context, accessToken string) (*UserSession, error) {
	userSession := new(UserSession)

	err := userSessionRepository.Database.
		NewSelect().
		Model(userSession).
		Where("access_token = ?", accessToken).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserSessionNotFound
		}

		return nil, err
	}

	return userSession, nil
}

func (userSessionRepository *UserSessionRepository) DeleteUserSessionByAccessToken(ctx context.Context, accessToken string) error {
	result, err := userSessionRepository.Database.
		NewDelete().
		Model(&UserSession{}).
		Where("access_token = ?", accessToken).
		Exec(ctx)

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrUserSessionNotFound
	}

	return err
}
