package messaging

import (
	"context"
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func PublishMessage[T EventData](ctx context.Context, jetStream nats.JetStream, message Message[T]) (*nats.PubAck, error) {
	tracer := otel.Tracer("apps-hosting.com/messaging")

	ctx, span := tracer.Start(ctx, message.Event)
	defer span.End()

	b_message, err := json.Marshal(message)
	if err != nil {
		log.Fatal("failed to encode json to bytes:", err)
	}

	m := &nats.Msg{
		Subject: message.Event,
		Data:    b_message,
		Header:  nats.Header{},
	}

	carrier := propagation.HeaderCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	for k, vs := range carrier {
		for _, v := range vs {
			m.Header.Add(k, v)
		}
	}

	return jetStream.PublishMsg(m)
}
