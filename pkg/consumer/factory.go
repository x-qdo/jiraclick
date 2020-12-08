package consumer

import (
	"x-qdo/jiraclick/pkg/contract"
	"x-qdo/jiraclick/pkg/provider"
	"x-qdo/jiraclick/pkg/provider/clickup"
	"x-qdo/jiraclick/pkg/publisher"
)

type inputBody struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Data        struct {
		Payload string `json:"payload"`
	} `json:"data"`
}

func MakeAction(
	key contract.RoutingKey,
	jira *provider.JiraClient,
	clickup *clickup.ClickUpAPIClient,
	publisher *publisher.EventPublisher,
) (contract.Action, error) {
	var (
		action contract.Action
		err    error
	)

	switch key {
	case contract.TaskCreateClickUp:
		action, err = NewTaskCreateClickupAction(clickup, publisher)
	case contract.TaskCreateJira:
		action, err = NewTaskCreateJiraAction(jira, publisher)
	case contract.TaskUpdateClickUp:
		action, err = NewTaskUpdateClickupAction(clickup, publisher)
	}

	if err != nil {
		return nil, err
	}

	return action, nil
}
