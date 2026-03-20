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
	conn *amqp091.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // SimpleQueueType is an "enum" to represent "durable" or "transient"
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

func SubscribeJSON[T any](
	conn *amqp091.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T),
) error {
	connCh, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return fmt.Errorf("QUEUE DECLARE/BIND ERROR: %s", err)
	}

	deliveries, err := connCh.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	// unsure if errs go anywhere, using handling method I looked up just blocks
	go func() error {
		for d := range deliveries {
			var data T
			err = json.Unmarshal(d.Body, &data)
			if err != nil {
				return fmt.Errorf("JSON ERROR: %s", err)
			}
			handler(data)
			d.Ack(false)
			if err != nil {
				return fmt.Errorf("ACKNOWLEDGE ERROR: %s", err)
			}
		}
		return nil
	}()

	return nil
}
