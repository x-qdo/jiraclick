package consumer

import (
	"x-qdo/jiraclick/pkg/config"
	"x-qdo/jiraclick/pkg/contract"
	"x-qdo/jiraclick/pkg/provider"
	"x-qdo/jiraclick/pkg/publisher"
)

func MakeAction(key contract.RoutingKey, cfg *config.Config, clickup *provider.ClickUpAPIClient, queueProvider *provider.RabbitChannel) (contract.Action, error) {
	var (
		action contract.Action
		err    error
	)

	p, err := publisher.NewEventPublisher(queueProvider)
	if err != nil {
		return nil, err
	}

	switch key {
	case contract.TaskCreateClickUp:
		action, err = NewTaskCreateClickupAction(cfg, clickup, p)
	}

	if err != nil {
		return nil, err
	}

	return action, nil
}
