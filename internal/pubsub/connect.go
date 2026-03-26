package pubsub

import (
	"fmt"

	"github.com/OminousOmelet/learn-pub-sub-starter/internal/routing"
	"github.com/rabbitmq/amqp091-go"
)

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

	table := amqp091.Table{"x-dead-letter-exchange": routing.ExchangePerilDeadLetter}
	queue, err := connCh.QueueDeclare(queueName, durable, autoDelete, exclusive, false, table)
	if err != nil {
		return nil, amqp091.Queue{}, fmt.Errorf("pubsub error: failed to declare queue: %s", err)
	}

	err = connCh.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, amqp091.Queue{}, fmt.Errorf("pubsub error: failed to bind queue to exchange: %s", err)
	}

	return connCh, queue, nil
}
