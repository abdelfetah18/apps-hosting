package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type ServiceLogger struct {
	Service Service
}

func NewServiceLogger(service Service) ServiceLogger {
	return ServiceLogger{Service: service}
}

func (logger ServiceLogger) LogInfo(message string) {
	logToStd(message, LogLabels{
		Type:    LogTypeService,
		Level:   LevelInfo,
		Service: logger.Service,
	})
}

func (logger ServiceLogger) LogInfoF(format string, a ...any) {
	logToStd(fmt.Sprintf(format, a...), LogLabels{
		Type:    LogTypeService,
		Level:   LevelInfo,
		Service: logger.Service,
	})
}

func (logger ServiceLogger) LogError(message string) {
	logToStd(message, LogLabels{
		Type:    LogTypeService,
		Level:   LevelError,
		Service: logger.Service,
	})
}

func (logger ServiceLogger) LogErrorF(format string, a ...any) {
	logToStd(fmt.Sprintf(format, a...), LogLabels{
		Type:    LogTypeService,
		Level:   LevelError,
		Service: logger.Service,
	})
}

type UserAppLogger struct {
	AppID  string
	UserID string
	Stage  Stage
}

func NewUserAppLogger(appID, userID string, stage Stage) UserAppLogger {
	return UserAppLogger{
		AppID:  appID,
		UserID: userID,
		Stage:  stage,
	}
}

func (logger UserAppLogger) LogInfo(message string) {
	logToStd(message, LogLabels{
		Type:   LogTypeUser,
		Level:  LevelInfo,
		Stage:  logger.Stage,
		UserID: logger.UserID,
		AppID:  logger.AppID,
	})
}

func (logger UserAppLogger) LogInfoF(format string, a ...any) {
	logToStd(fmt.Sprintf(format, a...), LogLabels{
		Type:   LogTypeUser,
		Level:  LevelInfo,
		Stage:  logger.Stage,
		UserID: logger.UserID,
		AppID:  logger.AppID,
	})
}

func (logger UserAppLogger) LogError(message string) {
	logToStd(message, LogLabels{
		Type:   LogTypeUser,
		Level:  LevelError,
		Stage:  logger.Stage,
		UserID: logger.UserID,
		AppID:  logger.AppID,
	})
}

func (logger UserAppLogger) LogErrorF(format string, a ...any) {
	logToStd(fmt.Sprintf(format, a...), LogLabels{
		Type:   LogTypeUser,
		Level:  LevelError,
		Stage:  logger.Stage,
		UserID: logger.UserID,
		AppID:  logger.AppID,
	})
}

func (logger ServiceLogger) Write(p []byte) (n int, err error) {
	logger.LogInfo(string(bytes.TrimSpace(p)))
	return len(p), nil
}

func (logger UserAppLogger) Write(p []byte) (n int, err error) {
	logger.LogInfo(string(bytes.TrimSpace(p)))
	return len(p), nil
}

func logToStd(message string, labels LogLabels) {
	entry := LogEntry{
		Timestamp: time.Now().UTC(),
		Message:   message,
		Labels:    labels.ToMap(),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to serialize log: %v\n", err)
		return
	}

	switch labels.Level {
	case LevelError, LevelFatal:
		fmt.Fprintln(os.Stderr, string(data))
	default:
		fmt.Fprintln(os.Stdout, string(data))
	}

	if labels.Level == LevelFatal {
		os.Exit(1)
	}
}
