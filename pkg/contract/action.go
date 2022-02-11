package contract

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Action interface {
	ProcessAction(ctx context.Context, delivery amqp.Delivery) error
}
