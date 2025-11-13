package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func SubscribeToEvent[T EventData](serviceName string, jetStream nats.JetStream, event AppEvent, handler EventHandler[T]) {
	_, err := jetStream.Subscribe(event, func(msg *nats.Msg) {
		message := Message[T]{}
		err := json.Unmarshal(msg.Data, &message)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		carrier := propagation.HeaderCarrier{}
		for k, vs := range msg.Header {
			for _, v := range vs {
				carrier.Set(k, v)
			}
		}

		log.Println("carrier:", carrier)

		ctx := otel.GetTextMapPropagator().Extract(context.Background(), carrier)

		tracer := otel.Tracer("apps-hosting.com/messaging")
		ctx, span := tracer.Start(ctx, event)
		defer span.End()

		message.Context = ctx

		handler(message)
	},
		nats.Durable(serviceName),
		nats.AckWait(5*time.Minute), // FIXME: build-service may take longer.
		nats.DeliverAll())

	if err != nil {
		fmt.Printf("SubscribeToEvent: error=%s\n", err.Error())
	} else {
		fmt.Printf("SubscribeToEvent: serviceName=%s, event=%s\n", serviceName, event)
	}
}
