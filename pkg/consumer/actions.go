package consumer

import (
	"github.com/astreter/amqpwrapper/v2"
	"x-qdo/jiraclick/pkg/contract"
	"x-qdo/jiraclick/pkg/provider/clickup"
	"x-qdo/jiraclick/pkg/provider/jira"
	"x-qdo/jiraclick/pkg/publisher"
)

var actionRoutingKeys = [3]contract.RoutingKey{
	contract.TaskCreateClickUp,
	contract.TaskCreateJira,
	contract.TaskUpdateClickUp,
}

type ActionsConsumer struct {
	queueProvider   *amqpwrapper.RabbitChannel
	clickupProvider *clickup.ConnectorPool
	jiraProvider    *jira.ConnectorPool
}

func NewActionsConsumer(
	jiraProvider *jira.ConnectorPool,
	queueProvider *amqpwrapper.RabbitChannel,
	clickup *clickup.ConnectorPool,
) (*ActionsConsumer, error) {
	if err := queueProvider.DefineExchange(contract.BRPActionsExchange, true); err != nil {
		return nil, err
	}

	return &ActionsConsumer{
		queueProvider:   queueProvider,
		clickupProvider: clickup,
		jiraProvider:    jiraProvider,
	}, nil
}

func (c *ActionsConsumer) SetUpListeners() error {
	p, err := publisher.NewEventPublisher(c.queueProvider)
	if err != nil {
		return err
	}

	for _, key := range actionRoutingKeys {
		action, err := MakeAction(key, c.jiraProvider, c.clickupProvider, p)
		if err != nil {
			return err
		}
		queueRoutingKey := string(key)
		err = c.queueProvider.SetUpConsumer(contract.BRPActionsExchange, queueRoutingKey, action.ProcessAction)
		if err != nil {
			return err
		}
	}

	return nil
}
