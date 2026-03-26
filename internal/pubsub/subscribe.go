package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int
type AckType int

const (
	Durable SimpleQueueType = iota
	Transient
)

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func SubscribeJSON[T any](
	conn *amqp091.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {
	return subscribe(conn, exchange, queueName, key, queueType, handler,
		func(data []byte) (T, error) {
			var jData T
			err := json.Unmarshal(data, &jData)
			return jData, err
		},
	)
}

func SubscribeGOB[T any](
	conn *amqp091.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {
	return subscribe(conn, exchange, queueName, key, queueType, handler,
		func(data []byte) (T, error) {
			var gData T
			var buff bytes.Buffer
			buff.Write(data)
			dec := gob.NewDecoder(&buff)
			err := dec.Decode(&gData)
			return gData, err
		},
	)
}

func subscribe[T any](
	conn *amqp091.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
	serializer func([]byte) (T, error),
) error {
	connCh, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return fmt.Errorf("QUEUE DECLARE/BIND ERROR: %s", err)
	}

	err = connCh.Qos(10, 0, true)
	if err != nil {
		return fmt.Errorf("channel QOS issue: %s", err)
	}

	deliveries, err := connCh.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("QUEUE CONSUMPTION ERROR: %s", err)
	}

	go func() {
		for d := range deliveries {
			data, err := serializer(d.Body)
			if err != nil {
				fmt.Printf("data serialization error: %s", err)
				continue
			}

			ack := handler(data)
			ackStr := ""

			switch ack {
			case Ack:
				ackStr = "Ack"
				d.Ack(false)
			case NackRequeue:
				ackStr = "NackRequeue"
				d.Nack(false, true)
			case NackDiscard:
				ackStr = "NackDiscard"
				d.Nack(false, false)
			}
			fmt.Printf("acktype '%s' succesfully processed.\n> ", ackStr)
		}
	}()

	return nil
}
