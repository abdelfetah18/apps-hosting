package handlers

import (
	"context"
	"encoding/json"
	"gateway/proto/user_service_pb"
	"net/http"

	"apps-hosting.com/messaging"

	"apps-hosting.com/logging"

	"google.golang.org/grpc/status"
)

type UserHandler struct {
	UserServiceClient user_service_pb.UserServiceClient
	Logger            logging.ServiceLogger
}

func NewUserHandler(userServiceClient user_service_pb.UserServiceClient, logger logging.ServiceLogger) UserHandler {
	return UserHandler{
		UserServiceClient: userServiceClient,
		Logger:            logger,
	}
}

func (handler *UserHandler) AuthHandler(w http.ResponseWriter, r *http.Request) {
	accessToken := r.Header.Get("Authorization")

	authResponse, err := handler.UserServiceClient.Auth(context.Background(), &user_service_pb.AuthRequest{AccessToken: accessToken})
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, http.StatusInternalServerError, status.Message())
		return
	}

	messaging.WriteSuccess(w, "Authentication passed", authResponse.User)
}

func (handler *UserHandler) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	signUpRequest := user_service_pb.SignUpRequest{}

	err := json.NewDecoder(r.Body).Decode(&signUpRequest)
	if err != nil {
		messaging.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	signUpResponse, err := handler.UserServiceClient.SignUp(context.Background(), &signUpRequest)
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, http.StatusInternalServerError, status.Message())
		return
	}

	messaging.WriteSuccess(w, "SignUp Successfully", signUpResponse.User)
}

func (handler *UserHandler) SignInHandler(w http.ResponseWriter, r *http.Request) {
	signInRequest := user_service_pb.SignInRequest{}

	err := json.NewDecoder(r.Body).Decode(&signInRequest)
	if err != nil {
		messaging.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	signInResponse, err := handler.UserServiceClient.SignIn(context.Background(), &signInRequest)
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, http.StatusInternalServerError, status.Message())
		return
	}

	messaging.WriteSuccess(w, "SignIn Successfully", signInResponse.UserSession)
}

func (handler *UserHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken := r.Header.Get("Authorization")
		authResponse, err := handler.UserServiceClient.Auth(context.Background(), &user_service_pb.AuthRequest{
			AccessToken: accessToken,
		})

		if err != nil {
			handler.Logger.LogError(err.Error())
			status, _ := status.FromError(err)
			messaging.WriteError(w, http.StatusInternalServerError, status.Message())
			return
		}

		q := r.URL.Query()
		q.Set("user_id", authResponse.User.Id)
		r.URL.RawQuery = q.Encode()

		next.ServeHTTP(w, r)
	})
}

func (handler *UserHandler) GithubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	userId := r.URL.Query().Get("user_id")

	exchangeGitHubCodeForTokenResponse, err := handler.UserServiceClient.ExchangeGitHubCodeForToken(context.Background(), &user_service_pb.ExchangeGitHubCodeForTokenRequest{
		Code:   code,
		UserId: userId,
	})

	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, http.StatusInternalServerError, status.Message())
		return
	}

	messaging.WriteSuccess(w, "Github Callback Successfully", exchangeGitHubCodeForTokenResponse)
}

func (handler *UserHandler) GetGithubRepositoriesHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("user_id")

	getGithubRepositoriesResponse, err := handler.UserServiceClient.GetGithubRepositories(context.Background(), &user_service_pb.GetGithubRepositoriesRequest{
		UserId: userId,
	})

	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, http.StatusInternalServerError, status.Message())
		return
	}

	messaging.WriteSuccess(w, "Github Repositories Fetched Successfully", getGithubRepositoriesResponse.GithubRepositories)
}
