package nats

import (
	"os"

	"apps-hosting.com/messaging"

	"apps-hosting.com/logging"

	"github.com/nats-io/nats.go"
)

type NatsService struct {
	JetStream nats.JetStream
	Logger    logging.ServiceLogger
}

func NewNatsService(logger logging.ServiceLogger) (*NatsService, error) {
	natsURL := os.Getenv("NATS_URL")
	natsClient, err := nats.Connect(natsURL)
	if err != nil {
		logger.LogError(err.Error())
		return nil, err
	}

	jetStream, err := natsClient.JetStream()
	if err != nil {
		logger.LogError(err.Error())
		return nil, err
	}

	_, err = jetStream.AddStream(&nats.StreamConfig{
		Name: "APP_EVENTS",
		Subjects: []string{
			messaging.AppCreated,
			messaging.AppReDeployRequested,
			messaging.AppExposed,
			messaging.AppDomainAssigned,
			messaging.AppDeleted,
		},
	})
	if err != nil {
		logger.LogError(err.Error())
		return nil, err
	}

	return &NatsService{
		JetStream: jetStream,
		Logger:    logger,
	}, nil
}

func (natsService *NatsService) SubscribeToEvents() {
	// messaging.SubscribeToEvent(NatsClient, messaging.BuildFailed, HandleBuildFailedEvent)
}
