package messaging

import (
	"time"

	"github.com/google/uuid"
)

type Group = string
type AppEvent = string
type Runtime = string

const (
	GroupDNSService    Group = "dns-service-group"
	GroupBuildService  Group = "build-service-group"
	GroupDeployService Group = "deploy-service-group"
)

const (
	AppCreated           AppEvent = "app.created"
	AppReDeployRequested AppEvent = "app.redeploy_requested"
	AppExposed           AppEvent = "app.exposed"
	AppDomainAssigned    AppEvent = "app.domain.assigned"
	AppDeleted           AppEvent = "app.deleted"

	BuildCompleted AppEvent = "build.completed"
	BuildFailed    AppEvent = "build.failed"

	DeployCompleted AppEvent = "deploy.completed"
	DeployFailed    AppEvent = "deploy.failed"
)

type Message[T EventData] struct {
	Id        string   `json:"id"`
	Event     AppEvent `json:"event"`
	Timestamp int64    `json:"timestamp"`
	Data      T        `json:"data"`
}

type AppCreatedData struct {
	RepoURL    string  `json:"repo_url"`
	AppName    string  `json:"app_name"`
	AppId      string  `json:"app_id"`
	UserId     string  `json:"user_id"`
	Runtime    Runtime `json:"runtime"`
	StartCMD   Runtime `json:"start_cmd"`
	BuildCMD   Runtime `json:"build_cmd"`
	DomainName string  `json:"domain_name"`
}

type AppReDeployRequestedData struct {
	RepoURL  string  `json:"repo_url"`
	AppName  string  `json:"app_name"`
	AppId    string  `json:"app_id"`
	Runtime  Runtime `json:"runtime"`
	StartCMD Runtime `json:"start_cmd"`
	BuildCMD Runtime `json:"build_cmd"`
	UserId   string  `json:"user_id"`
}

type BuildCompletedData struct {
	AppName    string `json:"app_name"`
	AppId      string `json:"app_id"`
	BuildId    string `json:"build_id"`
	ImageURL   string `json:"image_url"`
	Duration   int    `json:"duration"`
	DomainName string `json:"domain_name"`
}

type BuildFailedData struct {
	AppName string `json:"app_name"`
	AppId   string `json:"app_id"`
	BuildId string `json:"build_id"`
	Reason  string `json:"reason"`
}

type DeployCompletedData struct {
	AppName  string `json:"app_name"`
	DeployId string `json:"deploy_id"`
}

type AppExposedData struct {
	AppId    string `json:"app_id"`
	AppName  string `json:"app_name"`
	PublicIp string `json:"public_ip"`
}

type AppDomainAssignedData struct {
	AppId      string `json:"app_id"`
	DomainName string `json:"domain_name"`
	AppName    string `json:"app_name"`
}

type AppDeletedData struct {
	AppId   string `json:"app_id"`
	AppName string `json:"app_name"`
}

type ErrorEventData struct {
	AppName      string `json:"app_name"`
	ErrorMessage string `json:"error_message"`
}

type EventData interface {
	BuildCompletedData | DeployCompletedData | ErrorEventData | AppReDeployRequestedData | AppExposedData | AppDomainAssignedData | AppCreatedData | AppDeletedData | BuildFailedData
}

type EventHandler[T EventData] func(message Message[T])

func NewMessage[T EventData](event AppEvent, data T) Message[T] {
	return Message[T]{
		Id:        uuid.New().String(),
		Event:     event,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
}
