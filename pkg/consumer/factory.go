package consumer

import (
	"x-qdo/jiraclick/pkg/config"
	"x-qdo/jiraclick/pkg/contract"
)

type ActionFactory struct {
}

func (f *ActionFactory) Make(key contract.RoutingKey, cfg *config.Config) (contract.Action, error) {
	var (
		action contract.Action
		err    error
	)
	switch key {
	case contract.TaskCreate:
		action, err = NewTaskCreateClickupAction(cfg)
	}

	if err != nil {
		return nil, err
	}

	return action, nil
}
