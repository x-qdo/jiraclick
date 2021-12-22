package contract

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type Action interface {
	ProcessAction(delivery amqp.Delivery) error
}
