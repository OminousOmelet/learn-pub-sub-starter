package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/OminousOmelet/learn-pub-sub-starter/internal/routing"
	"github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	Durable = iota
	Transient
)

func PublishJSON[T any](ch *amqp091.Channel, exchange, key string, val T) error {
	valData, err := json.Marshal(val)
	ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp091.Publishing{
		ContentType: "application/json", Body: valData,
	})
	if err != nil {
		return fmt.Errorf("error calling Publish method on channel: %s", err)
	}

	return nil
}

func DeclareAndBind(
	conn *amqp091.Connection, exchange, queueName, key string, queueType SimpleQueueType, // SimpleQueueType is an "enum" type I made to represent "durable" or "transient"
) (*amqp091.Channel, amqp091.Queue, error) {
	connCh, err := conn.Channel()
	if err != nil {
		return nil, amqp091.Queue{}, fmt.Errorf("pubsub error: failed to open channel: %s", err)
	}

	fmt.Println("pubsub: Connection Successful")
	durable, autoDelete, exclusive := false, false, false
	if queueType == Durable {
		durable = true
	} else {
		autoDelete, exclusive = true, true
	}
	queue, err := connCh.QueueDeclare(queueName, durable, autoDelete, exclusive, false, nil)
	if err != nil {
		return nil, amqp091.Queue{}, fmt.Errorf("pubsub error: failed to declare queue: %s", err)
	}

	err = connCh.QueueBind(queueName, routing.PauseKey, exchange, false, nil)
	if err != nil {
		return nil, amqp091.Queue{}, fmt.Errorf("pubsub error: failed to bind queue to exchange: %s", err)
	}

	return connCh, queue, nil
}
