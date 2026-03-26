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
	_, err := subscribe("json", conn, exchange, queueName, key, queueType, handler)
	if err != nil {
		return fmt.Errorf("SUB FAIL: %s", err)
	}
	return nil
}

func SubscribeGOB[T any](
	conn *amqp091.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) (*amqp091.Channel, error) {
	connCh, err := subscribe("gob", conn, exchange, queueName, key, queueType, handler)
	if err != nil {
		return &amqp091.Channel{}, fmt.Errorf("SUB FAIL: %s", err)
	}

	return connCh, nil
}

func subscribe[T any](
	msgFormat string,
	conn *amqp091.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
) (*amqp091.Channel, error) {
	connCh, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return &amqp091.Channel{}, fmt.Errorf("QUEUE DECLARE/BIND ERROR: %s", err)
	}

	deliveries, err := connCh.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return &amqp091.Channel{}, err
	}

	go func() {
		for d := range deliveries {
			var data T
			if msgFormat == "json" {
				err = json.Unmarshal(d.Body, &data)
				if err != nil {
					fmt.Printf("JSON ERROR: %s", err)
					continue
				}
			} else if msgFormat == "gob" {
				var buff bytes.Buffer
				buff.Write(d.Body)
				dec := gob.NewDecoder(&buff)
				err = dec.Decode(&data)
				if err != nil {
					fmt.Printf("GOB ERROR: %s", err)
					continue
				}
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

	return connCh, nil
}
