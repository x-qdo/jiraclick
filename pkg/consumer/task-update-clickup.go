package consumer

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"x-qdo/jiraclick/pkg/contract"
	"x-qdo/jiraclick/pkg/model"
	"x-qdo/jiraclick/pkg/provider"
	"x-qdo/jiraclick/pkg/publisher"
)

type TaskUpdateClickupAction struct {
	client    *provider.ClickUpAPIClient
	publisher *publisher.EventPublisher
}

func NewTaskUpdateClickupAction(clickup *provider.ClickUpAPIClient, p *publisher.EventPublisher) (contract.Action, error) {
	return &TaskUpdateClickupAction{
		client:    clickup,
		publisher: p,
	}, nil
}

func (a *TaskUpdateClickupAction) ProcessAction(delivery amqp.Delivery) error {

	var (
		inputBody inputBody
		payload   model.TaskPayload
	)

	err := json.Unmarshal(delivery.Body, &inputBody)
	if err != nil {
		return errors.Wrap(err, "Can't unmarshall task body")
	}

	err = json.Unmarshal([]byte(inputBody.Data.Payload), &payload)
	if err != nil {
		return errors.Wrap(err, "Can't unmarshall task body")
	}

	request, err := a.generateTaskRequest(payload)
	if err != nil {
		return errors.Wrap(err, "Can't create task request")
	}

	err = a.client.UpdateTask(payload.ClickupID, request)
	if err != nil {
		return errors.Wrap(err, "Can't create a task in ClickUp")
	}

	return nil
}

func (a *TaskUpdateClickupAction) generateTaskRequest(payload model.TaskPayload) (*provider.PutClickUpTaskRequest, error) {
	request := new(provider.PutClickUpTaskRequest)

	request.Name = payload.Title
	request.Description = payload.Description + "\n\n" + payload.AC
	request.AddCustomField(provider.RequestedBy, payload.SlackReporter)
	request.AddCustomField(provider.SlackLink, payload.Details["slack"])
	request.AddCustomField(provider.JiraLink, payload.JiraUrl)

	return request, nil
}
