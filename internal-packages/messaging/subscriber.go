package messaging

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

func SubscribeToEvent[T EventData](serviceName string, jetStream nats.JetStream, event AppEvent, handler EventHandler[T]) {
	_, err := jetStream.Subscribe(event, func(msg *nats.Msg) {
		message := Message[T]{}
		err := json.Unmarshal(msg.Data, &message)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
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
