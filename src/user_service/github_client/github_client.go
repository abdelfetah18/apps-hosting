package githubclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"apps-hosting.com/logging"

	"github.com/google/go-github/v74/github"
)

type GithubOAuth struct {
	AccessToken           string `json:"access_token,omitempty"`
	Scope                 string `json:"scope,omitempty"`
	TokenType             string `json:"token_type,omitempty"`
	RefreshToken          string `json:"refresh_token,omitempty"`
	ExpiresIn             int64  `json:"expires_in,omitempty"`
	RefreshTokenExpiresIn int64  `json:"refresh_token_expires_in,omitempty"`
}

type GithubClient struct {
	Client *github.Client
	Logger logging.ServiceLogger
}

var clientId = os.Getenv("GITHUB_CLIENT_ID")
var clientSecret = os.Getenv("GITHUB_CLIENT_SECRET")

func NewGithubClient(accessToken *string, logger logging.ServiceLogger) GithubClient {
	if accessToken != nil {
		return GithubClient{
			Client: github.NewClient(nil).WithAuthToken(*accessToken),
			Logger: logger,
		}
	}

	return GithubClient{
		Client: github.NewClient(nil),
		Logger: logger,
	}
}

func (githubClient *GithubClient) RefreshToken(refreshToken string) (*GithubOAuth, error) {
	url := fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&refresh_token=%s&grant_type=refresh_token", clientId, clientSecret, refreshToken)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		githubClient.Logger.LogError(err.Error())
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		githubClient.Logger.LogError(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	response := GithubOAuth{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		githubClient.Logger.LogError(err.Error())
		return nil, err
	}

	return &response, nil
}

func (githubClient *GithubClient) Auth(code string) (*GithubOAuth, error) {
	url := fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s", clientId, clientSecret, code)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		githubClient.Logger.LogError(err.Error())
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		githubClient.Logger.LogError(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("response code was: " + strconv.Itoa(resp.StatusCode))
	}

	response := GithubOAuth{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		githubClient.Logger.LogError(err.Error())
		return nil, err
	}

	return &response, nil
}

func (githubClient *GithubClient) GetUserRepositories() ([]*github.Repository, error) {
	githubRepositories, _, err := githubClient.Client.Repositories.ListByAuthenticatedUser(context.Background(), &github.RepositoryListByAuthenticatedUserOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	})
	if err != nil {
		githubClient.Logger.LogError(err.Error())
		return nil, err
	}

	return githubRepositories, nil
}
