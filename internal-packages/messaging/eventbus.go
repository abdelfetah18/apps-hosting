package messaging

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"apps-hosting.com/messaging/proto/events_pb"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/protobuf/proto"
)

type EventHandler func(ctx context.Context, message *events_pb.Message)

type EventBus struct {
	serviceName string
	conn        *nats.Conn
	jetStream   nats.JetStreamContext
}

func NewEventBus(serviceName string, natsURL string, streamName events_pb.StreamName, subjects []events_pb.EventName) (*EventBus, error) {
	natsConnection, err := nats.Connect(natsURL)
	if err != nil {
		return nil, err
	}

	jetStream, err := natsConnection.JetStream()
	if err != nil {
		return nil, err
	}

	_subjects := []string{}
	for _, subject := range subjects {
		_subjects = append(_subjects, getEventName(subject))
	}

	_, err = jetStream.AddStream(&nats.StreamConfig{
		Name:     events_pb.StreamName_name[int32(streamName)],
		Subjects: _subjects,
	})
	if err != nil {
		return nil, err
	}

	return &EventBus{serviceName: serviceName, conn: natsConnection, jetStream: jetStream}, nil
}

func (e *EventBus) Subscribe(eventName events_pb.EventName, handler EventHandler) error {
	_, err := e.jetStream.Subscribe(getEventName(eventName), func(msg *nats.Msg) {
		//	1. validate the message
		message := events_pb.Message{}
		err := proto.Unmarshal(msg.Data, &message)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		// 	2. initiate a new tracing span
		carrier := propagation.HeaderCarrier{}
		for k, vs := range msg.Header {
			for _, v := range vs {
				carrier.Set(k, v)
			}
		}

		ctx := otel.GetTextMapPropagator().Extract(context.Background(), carrier)
		tracer := otel.Tracer("apps-hosting.com/messaging")
		ctx, span := tracer.Start(ctx, fmt.Sprintf("handle %s", getEventName(eventName)))
		defer span.End()

		//	3. call the handler
		handler(ctx, &message)
	},
		nats.Durable(fmt.Sprintf("%s-%s", e.serviceName, strings.ReplaceAll(getEventName(eventName), ".", "-"))),
		nats.AckWait(5*time.Minute), // FIXME: build-service may take longer.
		nats.DeliverAll())

	return err
}

func (e *EventBus) Publish(ctx context.Context, eventName events_pb.EventName, data *events_pb.EventData) error {
	tracer := otel.Tracer("apps-hosting.com/messaging")

	ctx, span := tracer.Start(ctx, fmt.Sprintf("publish %s", getEventName(eventName)))
	defer span.End()

	message := events_pb.Message{
		Id:        uuid.New().String(),
		EventName: eventName,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	b_message, err := proto.Marshal(&message)
	if err != nil {
		log.Fatal("failed to encode json to bytes:", err)
	}

	m := &nats.Msg{
		Subject: getEventName(eventName),
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

	_, err = e.jetStream.PublishMsg(m)
	return err
}

func getEventName(eventName events_pb.EventName) string {
	switch eventName {
	// App Events
	case events_pb.EventName_APP_CREATED:
		return "app.created"
	case events_pb.EventName_APP_DELETED:
		return "app.deleted"

	// Build Events
	case events_pb.EventName_BUILD_COMPLETED:
		return "build.completed"
	case events_pb.EventName_BUILD_FAILED:
		return "build.failed"

	// Deploy Events
	case events_pb.EventName_DEPLOY_COMPLETED:
		return "deploy.completed"
	case events_pb.EventName_DEPLOY_FAILED:
		return "deploy.failed"

	// Project Events
	case events_pb.EventName_PROJECT_DELETED:
		return "project.deleted"

	default:
		return "unknown"
	}
}
