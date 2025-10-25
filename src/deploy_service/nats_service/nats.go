package nats_service

import (
	"apps-hosting.com/messaging"

	"github.com/nats-io/nats.go"
)

type NatsClient struct {
	JetStream   nats.JetStream
	NatsHandler NatsHandler
}

func NewNatsClient(jetStream nats.JetStream, natsHandler NatsHandler) NatsClient {
	return NatsClient{
		JetStream:   jetStream,
		NatsHandler: natsHandler,
	}
}

func (client *NatsClient) SubscribeToEvents() {
	messaging.SubscribeToEvent("deploy-service", client.JetStream, messaging.BuildCompleted, client.NatsHandler.HandleBuildCompletedEvent)
	messaging.SubscribeToEvent("deploy-service", client.JetStream, messaging.AppDeleted, client.NatsHandler.HandleAppDeletedEvent)
}
