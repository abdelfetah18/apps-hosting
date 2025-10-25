package logging

import "time"

type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
	LevelFatal LogLevel = "fatal"
)

type LogEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Message   string            `json:"message"`
	Labels    map[string]string `json:"labels"`
}

type LogType string

const (
	LogTypeUser    LogType = "user"
	LogTypeService LogType = "service"
)

type Stage string

const (
	StageBuild   Stage = "build"
	StageRuntime Stage = "runtime"
)

type Service string

const (
	ServiceBuild   Service = "build_service"
	ServiceDeploy  Service = "deploy_service"
	ServiceApp     Service = "app_service"
	ServiceUser    Service = "user_service"
	ServiceGateway Service = "gateway_service"
	ServiceLog     Service = "log_service"
	ServiceProject Service = "project_service"
)

type LogLabels struct {
	Type    LogType  `json:"type"`              // "user" or "platform"
	Level   LogLevel `json:"level"`             // "info", "error", "debug", etc.
	Stage   Stage    `json:"stage,omitempty"`   // "build" or "runtime" (user only)
	UserID  string   `json:"user_id,omitempty"` // for user logs
	AppID   string   `json:"app_id,omitempty"`  // for user logs
	Service Service  `json:"service,omitempty"` // for platform logs
}

func (l LogLabels) ToMap() map[string]string {
	m := map[string]string{
		"type":  string(l.Type),
		"level": string(l.Level),
	}

	if l.Stage != "" {
		m["stage"] = string(l.Stage)
	}
	if l.UserID != "" {
		m["user_id"] = l.UserID
	}
	if l.AppID != "" {
		m["app_id"] = l.AppID
	}
	if l.Service != "" {
		m["service"] = string(l.Service)
	}

	return m
}
