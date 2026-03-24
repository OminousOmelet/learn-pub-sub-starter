package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp091.Channel, exchange, key string, val T) error {
	valData, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("error publishing JSON: %s", err)
	}

	// .....what??
	return ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp091.Publishing{
		ContentType: "application/json", Body: valData,
	})
}

func PublishGOB[T any](ch *amqp091.Channel, exchange, key string, val T) error {
	var buff bytes.Buffer
	dec := gob.NewEncoder(&buff)
	err := dec.Encode(&val)
	if err != nil {
		return fmt.Errorf("error publishing GOB: %s", err)
	}

	// .....the fuck??
	return ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp091.Publishing{
		ContentType: "application/json", Body: buff.Bytes(),
	})
}
