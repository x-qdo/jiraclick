package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"sync"
	"time"
	"x-qdo/jiraclick/pkg/config"
)

type RabbitChannel struct {
	ctx             context.Context
	waitGroup       *sync.WaitGroup
	conn            *amqp.Connection
	channel         *amqp.Channel
	exchangeName    string
	url             string
	errorConnection chan *amqp.Error
	notifyConfirm   chan amqp.Confirmation
	closed          bool
	consumers       map[string]consumer
	logger          logrus.FieldLogger
}

type messageListener func(delivery amqp.Delivery) error

type consumer struct {
	name         string
	routingKey   string
	exchangeName string
	callback     messageListener
}

func NewRabbitChannel(ctx context.Context, wg *sync.WaitGroup, cfg *config.Config, logger logrus.FieldLogger) (*RabbitChannel, error) {
	ch := new(RabbitChannel)

	url := cfg.RabbitMQ.URL
	logger.Debug("RabbitMQ.URL: ", url)
	if url == "" {
		logger.Info("RabbitMQ.URL not found, building from components")
		url = "amqp://" + cfg.RabbitMQ.User + ":" + cfg.RabbitMQ.Password + "@" + cfg.RabbitMQ.Host + ":" + cfg.RabbitMQ.Port
		if cfg.RabbitMQ.Vhost != "" {
			url = url + cfg.RabbitMQ.Vhost
		}
		logger.Debug("RabbitMQ.URL: ", url)
	}

	ch.url = url

	err := ch.connect()
	if err != nil {
		return nil, err
	}
	go ch.reconnect()

	ch.ctx = ctx
	ch.waitGroup = wg
	ch.consumers = make(map[string]consumer)

	return ch, nil
}

func (ch *RabbitChannel) DefineExchange(exchangeName string, isAlreadyExist bool) error {
	var err error
	if isAlreadyExist {
		err = ch.channel.ExchangeDeclarePassive(
			exchangeName, // name
			"topic",      // type
			true,         // durable
			false,        // auto-deleted
			false,        // internal
			false,        // no-wait
			nil,          // arguments
		)
	} else {
		err = ch.channel.ExchangeDeclare(
			exchangeName, // name
			"topic",      // type
			true,         // durable
			false,        // auto-deleted
			false,        // internal
			false,        // no-wait
			nil,          // arguments
		)
	}
	if err != nil {
		return fmt.Errorf("RabbitMQ: failed to declare an exchange: %s", err.Error())
	}

	ch.logger.Debug("RabbitMQ: exchange `" + exchangeName + "` is declared")
	return nil
}

func (ch *RabbitChannel) Publish(message interface{}, exchangeName, routingKey string, safeMode bool) error {
	ch.waitGroup.Add(1)
	defer ch.waitGroup.Done()
	body, err := json.Marshal(message)
	if err != nil {
		ch.logger.Error(fmt.Errorf("RabbitMQ: failed to encode message: %w", err).Error())
		return err
	}

	if ch.closed {
		return errors.New("rabbitMQ: failed to publish a message: connection is lost")
	}

	for {
		err = ch.channel.Publish(
			exchangeName, // exchange
			routingKey,   // routing key
			false,        // mandatory
			false,        // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        body,
			})

		if err != nil {
			if err == amqp.ErrClosed {
				continue
			}
			ch.logger.Error(fmt.Errorf("RabbitMQ: failed to publish a message: %w", err).Error())
			return err
		}
		if safeMode {
			select {
			case confirm := <-ch.notifyConfirm:
				if confirm.Ack {
					return nil
				}
			case <-time.After(3 * time.Second):
				return errors.New("rabbitMQ: failed to publish a message: delivery confirmation is not received")
			}
		}
		return nil
	}
}

func (ch *RabbitChannel) SetUpConsumer(exchangeName, name, routingKey string, callback messageListener) error {
	q, err := ch.channel.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		err = fmt.Errorf("RabbitMQ: failed to declare a queue: %w", err)
		ch.logger.Error(err.Error())
		return err
	}

	err = ch.channel.QueueBind(
		q.Name,       // queue name
		routingKey,   // routing key
		exchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		err = fmt.Errorf("RabbitMQ: failed to bind a queue: %w", err)
		ch.logger.Error(err.Error())
		return err
	}

	msgChannel, err := ch.channel.Consume(
		q.Name, // queue
		name,   // consumer
		false,  // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	if err != nil {
		err = fmt.Errorf("RabbitMQ: failed to register a consumer: %w", err)
		ch.logger.Error(err.Error())
		return err
	}

	ch.consumers[name] = consumer{
		name:         name,
		routingKey:   routingKey,
		exchangeName: exchangeName,
		callback:     callback,
	}

	go ch.listenQueue(routingKey, msgChannel, callback)

	return nil
}

func (ch *RabbitChannel) Close() error {
	ch.closed = true
	_ = ch.channel.Close()
	err := ch.conn.Close()
	if err != nil {
		return fmt.Errorf("RabbitMQ: failed to close the connection: %w", err)
	}
	ch.logger.Debug("RabbitMQ: Connection is closed")
	return nil
}

func (ch *RabbitChannel) IsAlive() bool {
	return !ch.conn.IsClosed()
}

func (ch *RabbitChannel) connect() error {
	conn, err := amqp.Dial(ch.url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %s", err.Error())
	}
	ch.conn = conn
	ch.errorConnection = make(chan *amqp.Error)
	ch.conn.NotifyClose(ch.errorConnection)
	ch.logger.Debug("RabbitMQ: Connection is established")

	ch.channel, err = conn.Channel()
	if err != nil {
		return fmt.Errorf("RabbitMQ: failed to open a channel: %s", err.Error())
	}

	err = ch.channel.Confirm(false)
	if err != nil {
		return fmt.Errorf("RabbitMQ: %s", err.Error())
	}

	ch.notifyConfirm = make(chan amqp.Confirmation)
	ch.channel.NotifyPublish(ch.notifyConfirm)

	err = ch.channel.Qos(3, 0, true)
	if err != nil {
		return fmt.Errorf("RabbitMQ: failed to set QoS of a channel: %s", err.Error())
	}

	ch.logger.Debug("RabbitMQ: Channel is opened")
	return nil
}

func (ch *RabbitChannel) reconnect() {
	for {
		errorConnection := <-ch.errorConnection
		if !ch.closed {
			ch.logger.Error(fmt.Errorf("RabbitMQ: service tries to reconnect: %w", errorConnection).Error())

			err := ch.connect()
			if err != nil {
				ch.logger.Error(err.Error())
				panic(err)
			}
			ch.recoverConsumers()
		} else {
			return
		}
	}
}

func (ch *RabbitChannel) recoverConsumers() {
	for _, consumer := range ch.consumers {
		err := ch.SetUpConsumer(consumer.exchangeName, consumer.name, consumer.routingKey, consumer.callback)
		if err != nil {
			ch.logger.Error(err.Error())
		}
	}
}

func (ch *RabbitChannel) listenQueue(routingKey string, msgChannel <-chan amqp.Delivery, callback messageListener) {
	ch.logger.Debugf("listener for queue %s is in action", routingKey)
	defer ch.logger.Debugf("listener %s is stopped", routingKey)
	ch.waitGroup.Add(1)
	defer ch.waitGroup.Done()

	done := false
	ctx, _ := context.WithCancel(ch.ctx)

	for {
		select {
		case delivery, ok := <-msgChannel:
			if !ok {
				if _, _, err := ch.channel.Get(routingKey, true); err != nil {
					ch.logger.Error(err)
				}
				return
			}
			if err := callback(delivery); err != nil {
				if err := delivery.Nack(false, false); err != nil {
					ch.logger.Error(fmt.Errorf("RabbitMQ: %s: message nacking failed: %w. Consumer is turned off", routingKey, err))
					return
				}
			}
			if err := delivery.Ack(false); err != nil {
				ch.logger.Error(fmt.Errorf("%s: acknowledger failed with an error: %w", routingKey, err))
			}

			if done && len(msgChannel) == 0 {
				return
			}
		case <-ctx.Done():
			if err := ch.channel.Cancel(routingKey, false); err != nil {
				ch.logger.Error(err)
			}
			done = true
		}
		if ch.IsAlive() != true {
			ch.logger.Debug(fmt.Sprintf("Consumer %s has faced with closed channel", routingKey))
			return
		}
	}
}
