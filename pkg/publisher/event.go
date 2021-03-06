package publisher

import (
	"fmt"

	"github.com/pkg/errors"

	"x-qdo/jiraclick/pkg/contract"
	"x-qdo/jiraclick/pkg/model"
	"x-qdo/jiraclick/pkg/provider"
)

type EventPublisher struct {
	queueProvider *provider.RabbitChannel
}

func NewEventPublisher(queueProvider *provider.RabbitChannel) (*EventPublisher, error) {
	if err := queueProvider.DefineExchange(contract.BRPEventsExchange, true); err != nil {
		return nil, err
	}

	return &EventPublisher{
		queueProvider: queueProvider,
	}, nil
}

func (p *EventPublisher) ClickUpTaskCreated(payload model.TaskPayload) error {
	routingKey := fmt.Sprintf(string(contract.TaskCreatedClickUpEvent), payload.SlackChannel)
	if err := p.queueProvider.Publish(payload, contract.BRPEventsExchange, routingKey, true); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to send a %s to events queue", routingKey))
	}

	return nil
}

func (p *EventPublisher) JiraTaskCreated(payload model.TaskPayload) error {
	routingKey := fmt.Sprintf(string(contract.TaskCreatedJiraEvent), payload.SlackChannel)
	if err := p.queueProvider.Publish(payload, contract.BRPEventsExchange, routingKey, true); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to send a %s to events queue", routingKey))
	}

	return nil
}

func (p *EventPublisher) ClickUpTaskUpdated(payload model.TaskChanges, slackChannel string) error {
	routingKey := fmt.Sprintf(string(contract.TaskUpdatedClickUpEvent), slackChannel)
	if err := p.queueProvider.Publish(payload, contract.BRPEventsExchange, routingKey, true); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to send a %s to events queue", routingKey))
	}

	return nil
}
