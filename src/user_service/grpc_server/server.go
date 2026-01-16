package grpc_server

import (
	"context"
	"errors"
	"os"
	"strconv"
	"time"
	githubclient "user/github_client"
	"user/repositories"
	"user/utils"

	"apps-hosting.com/logging"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	user_service_pb "user/proto/user_service_pb"

	"github.com/golang-jwt/jwt/v5"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCUserServiceServer struct {
	user_service_pb.UnimplementedUserServiceServer
	UserRepository        repositories.UserRepository
	UserSessionRepository repositories.UserSessionRepository
	Logger                logging.ServiceLogger
}

func NewGRPCUserServiceServer(
	userRepository repositories.UserRepository,
	userSessionRepository repositories.UserSessionRepository,
	logger logging.ServiceLogger,
) *GRPCUserServiceServer {
	return &GRPCUserServiceServer{
		UserRepository:        userRepository,
		UserSessionRepository: userSessionRepository,
		Logger:                logger,
	}
}

func (server *GRPCUserServiceServer) Health(ctx context.Context, _ *user_service_pb.HealthRequest) (*user_service_pb.HealthResponse, error) {
	return &user_service_pb.HealthResponse{
		Status:  "success",
		Message: "OK",
	}, nil
}

func (server *GRPCUserServiceServer) Auth(ctx context.Context, authRequest *user_service_pb.AuthRequest) (*user_service_pb.AuthResponse, error) {
	span := trace.SpanFromContext(ctx)

	type AuthClaims struct {
		User repositories.User `json:"user"`
		jwt.RegisteredClaims
	}

	data := &AuthClaims{}
	_, err := jwt.ParseWithClaims(authRequest.AccessToken, data, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		switch {
		case errors.Is(err, jwt.ErrTokenMalformed):
			return nil, status.Error(codes.InvalidArgument, "token is malformed")
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, status.Error(codes.InvalidArgument, "token is expired")
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = server.UserSessionRepository.GetUserSessionByAccessToken(ctx, authRequest.AccessToken)

	if err == repositories.ErrUserSessionNotFound {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	user, err := server.UserRepository.GetUserById(ctx, data.User.Id)
	if err == repositories.ErrUserNotFound {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if user.GithubRefreshToken != "" && user.GithubRefreshTokenExpiresAt.Before(time.Now()) {
		user, err = server.UserRepository.UpdateUserGithub(ctx, user.Id, repositories.UpdateUserGithubParams{
			GithubAppInstalled: false,
			GithubAccessToken:  "",
			GithubRefreshToken: "",
		})

		if err != nil {
			span.SetAttributes(attribute.String("error", err.Error()))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if user.GithubAccessToken != "" && user.GithhubAccessTokenExpiresAt.Before(time.Now()) {
		githubClient := githubclient.NewGithubClient(nil, server.Logger)
		githubOAuth, err := githubClient.RefreshToken(user.GithubRefreshToken)
		if err != nil {
			span.SetAttributes(attribute.String("error", err.Error()))
			return nil, status.Error(codes.Internal, err.Error())
		}

		user, err = server.UserRepository.UpdateUserGithub(ctx, user.Id, repositories.UpdateUserGithubParams{
			GithubAppInstalled:          true,
			GithubAccessToken:           githubOAuth.AccessToken,
			GithubRefreshToken:          githubOAuth.RefreshToken,
			GithhubAccessTokenExpiresAt: time.Now().Add(time.Duration(githubOAuth.ExpiresIn) * time.Second),
			GithubRefreshTokenExpiresAt: time.Now().Add(time.Duration(githubOAuth.RefreshTokenExpiresIn) * time.Second),
		})

		if err != nil {
			span.SetAttributes(attribute.String("error", err.Error()))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	span.SetAttributes(attribute.String("user.id", user.Id))

	return &user_service_pb.AuthResponse{
		User: &user_service_pb.User{
			Id:                 user.Id,
			Username:           user.Username,
			Email:              user.Email,
			GithubAppInstalled: user.GithubAppInstalled,
			CreatedAt:          data.User.CreatedAt.String(),
		},
	}, nil
}

func (server *GRPCUserServiceServer) SignIn(ctx context.Context, signInRequest *user_service_pb.SignInRequest) (*user_service_pb.SignInResponse, error) {
	span := trace.SpanFromContext(ctx)

	user, err := server.UserRepository.GetUserByEmail(ctx, signInRequest.Email)

	if err == repositories.ErrUserNotFound {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	isValidPassword := utils.VerifyPassword(signInRequest.Password, user.Password)
	if !isValidPassword {
		return nil, status.Error(codes.InvalidArgument, "wrong password")
	}

	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": user,
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = server.UserSessionRepository.CreateUserSession(ctx, user.Id, repositories.CreateUserSessionParams{
		AccessToken: tokenString,
	})

	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user_service_pb.SignInResponse{
		UserSession: &user_service_pb.UserSession{
			AccessToken: tokenString,
			User: &user_service_pb.User{
				Id:                 user.Id,
				Username:           user.Username,
				Email:              user.Email,
				GithubAppInstalled: user.GithubAppInstalled,
				CreatedAt:          user.CreatedAt.String(),
			},
		},
	}, nil
}

func (server *GRPCUserServiceServer) SignUp(ctx context.Context, createUser *user_service_pb.SignUpRequest) (*user_service_pb.SignUpResponse, error) {
	span := trace.SpanFromContext(ctx)

	hashedPassword, err := utils.HashPassword(createUser.Password)
	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	createUser.Password = hashedPassword

	createdUser, err := server.UserRepository.CreateUser(ctx, repositories.CreateUserParams{
		Username: createUser.Username,
		Password: createUser.Password,
		Email:    createUser.Email,
	})

	if err == repositories.ErrUsernameInUse {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}

	if err == repositories.ErrEmailInUse {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}

	if err != nil {
		server.Logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	span.SetAttributes(attribute.String("user.id", createdUser.Id))

	return &user_service_pb.SignUpResponse{
		User: &user_service_pb.User{
			Id:        createdUser.Id,
			Email:     createdUser.Email,
			Username:  createdUser.Username,
			CreatedAt: createdUser.CreatedAt.String(),
		},
	}, nil
}

func (server *GRPCUserServiceServer) SignOut(ctx context.Context, signOutRequest *user_service_pb.SignOutRequest) (*user_service_pb.SignOutResponse, error) {
	span := trace.SpanFromContext(ctx)

	err := server.UserSessionRepository.DeleteUserSessionByAccessToken(ctx, signOutRequest.AccessToken)

	if err == repositories.ErrUserSessionNotFound {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user_service_pb.SignOutResponse{}, nil
}

func (server *GRPCUserServiceServer) GetGithubRepositories(ctx context.Context, getGithubRepositoriesRequest *user_service_pb.GetGithubRepositoriesRequest) (*user_service_pb.GetGithubRepositoriesResponse, error) {
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(attribute.String("user.id", getGithubRepositoriesRequest.UserId))

	user, err := server.UserRepository.GetUserById(ctx, getGithubRepositoriesRequest.UserId)
	if err == repositories.ErrUserNotFound {
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if err != nil {
		server.Logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if user.GithubAccessToken == "" {
		return nil, status.Error(codes.Unauthenticated, "Github account is not authenticated")
	}

	githubClient := githubclient.NewGithubClient(&user.GithubAccessToken, server.Logger)
	githubRepositories, err := githubClient.GetUserRepositories()
	if err != nil {
		server.Logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	safeString := func(p *string) string {
		if p == nil {
			return ""
		}
		return *p
	}

	// TODO: create a util function
	_githubRepositories := []*user_service_pb.GithubRepository{}
	githubRepositoriesIds := []string{}
	for _, githubRepository := range githubRepositories {
		_githubRepository := user_service_pb.GithubRepository{
			Id:            strconv.FormatInt(*githubRepository.ID, 10),
			Name:          safeString(githubRepository.Name),
			Description:   safeString(githubRepository.Description),
			DefaultBranch: safeString(githubRepository.DefaultBranch),
			GitUrl:        safeString(githubRepository.GitURL),
			Url:           safeString(githubRepository.URL),
			Visibility:    safeString(githubRepository.Visibility),
			CloneUrl:      safeString(githubRepository.CloneURL),
		}
		_githubRepositories = append(_githubRepositories, &_githubRepository)
		githubRepositoriesIds = append(githubRepositoriesIds, strconv.FormatInt(*githubRepository.ID, 10))
	}

	span.SetAttributes(
		attribute.StringSlice("github_repositories.ids", githubRepositoriesIds),
		attribute.Int("github_repositories.count", len(githubRepositories)),
	)

	return &user_service_pb.GetGithubRepositoriesResponse{
		GithubRepositories: _githubRepositories,
	}, nil
}

func (server *GRPCUserServiceServer) ExchangeGitHubCodeForToken(ctx context.Context, exchangeGitHubCodeForTokenRequest *user_service_pb.ExchangeGitHubCodeForTokenRequest) (*user_service_pb.ExchangeGitHubCodeForTokenResponse, error) {
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(attribute.String("user.id", exchangeGitHubCodeForTokenRequest.UserId))

	githubClient := githubclient.NewGithubClient(nil, server.Logger)
	githubOAuth, err := githubClient.Auth(exchangeGitHubCodeForTokenRequest.Code)

	if err != nil {
		server.Logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = server.UserRepository.UpdateUserGithub(ctx, exchangeGitHubCodeForTokenRequest.UserId, repositories.UpdateUserGithubParams{
		GithubAppInstalled:          true,
		GithubAccessToken:           githubOAuth.AccessToken,
		GithubRefreshToken:          githubOAuth.RefreshToken,
		GithhubAccessTokenExpiresAt: time.Now().Add(time.Duration(githubOAuth.ExpiresIn) * time.Second),
		GithubRefreshTokenExpiresAt: time.Now().Add(time.Duration(githubOAuth.RefreshTokenExpiresIn) * time.Second),
	})

	if err != nil {
		server.Logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user_service_pb.ExchangeGitHubCodeForTokenResponse{
		AccessToken:           githubOAuth.AccessToken,
		Scope:                 githubOAuth.Scope,
		TokenType:             githubOAuth.TokenType,
		RefreshToken:          githubOAuth.RefreshToken,
		ExpiresIn:             githubOAuth.ExpiresIn,
		RefreshTokenExpiresIn: githubOAuth.RefreshTokenExpiresIn,
	}, nil
}

func (server *GRPCUserServiceServer) GetGithubUserAccessToken(ctx context.Context, getGithubUserAccessTokenRequest *user_service_pb.GetGithubUserAccessTokenRequest) (*user_service_pb.GetGithubUserAccessTokenRespone, error) {
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(attribute.String("user.id", getGithubUserAccessTokenRequest.UserId))

	user, err := server.UserRepository.GetUserById(ctx, getGithubUserAccessTokenRequest.UserId)
	if err != nil {
		server.Logger.LogError(err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user_service_pb.GetGithubUserAccessTokenRespone{
		GithubUserAccessToken: user.GithubAccessToken,
	}, nil
}
