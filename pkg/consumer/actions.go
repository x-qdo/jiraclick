package consumer

import (
	"x-qdo/jiraclick/pkg/config"
	"x-qdo/jiraclick/pkg/contract"
	"x-qdo/jiraclick/pkg/provider"
)

var actionRoutingKeys = [1]contract.RoutingKey{
	contract.TaskCreate,
}

type ActionsConsumer struct {
	queueProvider *provider.RabbitChannel
	cfg           *config.Config
}

func NewActionsConsumer(cfg *config.Config, queueProvider *provider.RabbitChannel) (*ActionsConsumer, error) {

	if err := queueProvider.DefineExchange(contract.BRPActionsExchange, true); err != nil {
		return nil, err
	}

	return &ActionsConsumer{
		queueProvider: queueProvider,
		cfg:           cfg,
	}, nil
}

func (c *ActionsConsumer) SetUpListeners() error {
	var factory ActionFactory

	for _, key := range actionRoutingKeys {
		action, err := factory.Make(key, c.cfg)
		if err != nil {
			return err
		}
		queueRoutingKey := "t:*:" + string(key)
		err = c.queueProvider.SetUpConsumer(contract.BRPActionsExchange, string(key), queueRoutingKey, action.ProcessAction)
		if err != nil {
			return err
		}
	}

	return nil
}
