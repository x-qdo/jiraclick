package consumer

import (
	"x-qdo/jiraclick/pkg/config"
	"x-qdo/jiraclick/pkg/contract"
	"x-qdo/jiraclick/pkg/provider"
)

var actionRoutingKeys = [1]contract.RoutingKey{
	contract.TaskCreateClickUp,
}

type ActionsConsumer struct {
	queueProvider   *provider.RabbitChannel
	clickupProvider *provider.ClickUpAPIClient
	cfg             *config.Config
}

func NewActionsConsumer(cfg *config.Config, queueProvider *provider.RabbitChannel, clickup *provider.ClickUpAPIClient) (*ActionsConsumer, error) {

	if err := queueProvider.DefineExchange(contract.BRPActionsExchange, true); err != nil {
		return nil, err
	}

	return &ActionsConsumer{
		queueProvider:   queueProvider,
		clickupProvider: clickup,
		cfg:             cfg,
	}, nil
}

func (c *ActionsConsumer) SetUpListeners() error {
	for _, key := range actionRoutingKeys {
		action, err := MakeAction(key, c.cfg, c.clickupProvider, c.queueProvider)
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
