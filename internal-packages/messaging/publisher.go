package messaging

import (
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"
)

func PublishMessage[T EventData](jetStream nats.JetStream, message Message[T]) (*nats.PubAck, error) {
	b_message, err := json.Marshal(message)
	if err != nil {
		log.Fatal("failed to encode json to bytes:", err)
	}

	return jetStream.Publish(message.Event, b_message)
}
