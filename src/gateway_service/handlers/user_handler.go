package handlers

import (
	"encoding/json"
	"gateway/proto/user_service_pb"
	"net/http"

	"apps-hosting.com/messaging"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

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

	span := trace.SpanFromContext(r.Context())

	authResponse, err := handler.UserServiceClient.Auth(r.Context(), &user_service_pb.AuthRequest{AccessToken: accessToken})
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, http.StatusInternalServerError, status.Message())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	span.SetAttributes(attribute.String("user.id", authResponse.User.Id))
	messaging.WriteSuccess(w, "Authentication passed", authResponse.User)
}

func (handler *UserHandler) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	signUpRequest := user_service_pb.SignUpRequest{}
	span := trace.SpanFromContext(r.Context())
	err := json.NewDecoder(r.Body).Decode(&signUpRequest)
	if err != nil {
		messaging.WriteError(w, http.StatusBadRequest, err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	signUpResponse, err := handler.UserServiceClient.SignUp(r.Context(), &signUpRequest)
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, http.StatusInternalServerError, status.Message())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	span.SetAttributes(attribute.String("user.id", signUpResponse.User.Id))
	messaging.WriteSuccess(w, "SignUp Successfully", signUpResponse.User)
}

func (handler *UserHandler) SignInHandler(w http.ResponseWriter, r *http.Request) {
	span := trace.SpanFromContext(r.Context())
	signInRequest := user_service_pb.SignInRequest{}

	err := json.NewDecoder(r.Body).Decode(&signInRequest)
	if err != nil {
		messaging.WriteError(w, http.StatusBadRequest, err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	signInResponse, err := handler.UserServiceClient.SignIn(r.Context(), &signInRequest)
	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, http.StatusInternalServerError, status.Message())
		span.SetAttributes(attribute.String("error", err.Error()))
		return
	}

	span.SetAttributes(attribute.String("user.id", signInResponse.UserSession.User.Id))
	messaging.WriteSuccess(w, "SignIn Successfully", signInResponse.UserSession)
}

func (handler *UserHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken := r.Header.Get("Authorization")

		span := trace.SpanFromContext(r.Context())
		authResponse, err := handler.UserServiceClient.Auth(r.Context(), &user_service_pb.AuthRequest{
			AccessToken: accessToken,
		})

		if err != nil {
			handler.Logger.LogError(err.Error())
			status, _ := status.FromError(err)
			messaging.WriteError(w, http.StatusInternalServerError, status.Message())
			span.SetAttributes(attribute.String("error", err.Error()))
			return
		}

		span.SetAttributes(attribute.String("user.id", authResponse.User.Id))

		q := r.URL.Query()
		q.Set("user_id", authResponse.User.Id)
		r.URL.RawQuery = q.Encode()

		next.ServeHTTP(w, r)
	})
}

func (handler *UserHandler) GithubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	userId := r.URL.Query().Get("user_id")

	exchangeGitHubCodeForTokenResponse, err := handler.UserServiceClient.ExchangeGitHubCodeForToken(r.Context(), &user_service_pb.ExchangeGitHubCodeForTokenRequest{
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

	getGithubRepositoriesResponse, err := handler.UserServiceClient.GetGithubRepositories(r.Context(), &user_service_pb.GetGithubRepositoriesRequest{
		UserId: userId,
	})

	if err != nil {
		status, _ := status.FromError(err)
		messaging.WriteError(w, http.StatusInternalServerError, status.Message())
		return
	}

	messaging.WriteSuccess(w, "Github Repositories Fetched Successfully", getGithubRepositoriesResponse.GithubRepositories)
}
