package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
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
