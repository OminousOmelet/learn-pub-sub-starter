package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	Durable = iota
	Transient
)

type AckType int

const (
	Ack = iota
	NackRequeue
	NackDiscard
)

func PublishJSON[T any](ch *amqp091.Channel, exchange, key string, val T) error {
	valData, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("error calling Publish method on channel: %s", err)
	}

	// .....what??
	return ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp091.Publishing{
		ContentType: "application/json", Body: valData,
	})
}

func DeclareAndBind(
	conn *amqp091.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an "enum" to represent "durable" or "transient"
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

	err = connCh.QueueBind(queueName, key, exchange, false, nil)
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
	handler func(T) AckType,
) error {
	connCh, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return fmt.Errorf("QUEUE DECLARE/BIND ERROR: %s", err)
	}

	deliveries, err := connCh.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for d := range deliveries {
			var data T
			err = json.Unmarshal(d.Body, &data)
			if err != nil {
				fmt.Printf("JSON ERROR: %s", err)
				continue
			}
			ack := handler(data)
			ackStr := ""
			switch ack {
			case Ack:
				ackStr = "Ack"
				err = d.Ack(false)
				if err != nil {
					fmt.Printf("TYPE 'Ack' ERROR: %s", err)
					continue
				}
			case NackRequeue:
				ackStr = "NackRequeue"
				err = d.Nack(false, true)
				if err != nil {
					fmt.Printf("TYPE 'Nackreque' ERROR: %s", err)
					continue
				}
			case NackDiscard:
				ackStr = "NackDiscard"
				err = d.Nack(false, false)
				if err != nil {
					fmt.Printf("TYPE 'NackDiscard' ERROR: %s", err)
					continue
				}
			}
			fmt.Printf("acktype '%s' succesfully processed.\n> ", ackStr)
		}
	}()

	return nil
}
