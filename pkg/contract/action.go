package contract

import (
	"github.com/streadway/amqp"
)

type Action interface {
	ProcessAction(delivery amqp.Delivery) error
}
