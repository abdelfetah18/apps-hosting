package nats_service

import (
	"apps-hosting.com/messaging"

	"apps-hosting.com/logging"

	"github.com/nats-io/nats.go"
)

type NatsClient struct {
	JetStream   nats.JetStream
	NatsHandler NatsHandler
	Logger      logging.ServiceLogger
}

func NewNatsClient(jetStream nats.JetStream, natsHandler NatsHandler, logger logging.ServiceLogger) NatsClient {
	return NatsClient{
		JetStream:   jetStream,
		NatsHandler: natsHandler,
		Logger:      logger,
	}
}

func (natsService *NatsClient) SubscribeToEvents() {
	messaging.SubscribeToEvent(
		"build-service-app-created",
		natsService.JetStream,
		messaging.AppCreated,
		natsService.NatsHandler.HandleAppCreatedEvent,
	)

	messaging.SubscribeToEvent(
		"build-service-app-deleted",
		natsService.JetStream,
		messaging.AppDeleted,
		natsService.NatsHandler.HandleAppDeletedEvent,
	)
}
