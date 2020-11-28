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

type TaskCreateClickupAction struct {
	client    *provider.ClickUpAPIClient
	publisher *publisher.EventPublisher
}

func NewTaskCreateClickupAction(clickup *provider.ClickUpAPIClient, p *publisher.EventPublisher) (contract.Action, error) {
	return &TaskCreateClickupAction{
		client:    clickup,
		publisher: p,
	}, nil
}

func (a *TaskCreateClickupAction) ProcessAction(delivery amqp.Delivery) error {

	var (
		inputBody inputBody
		response  *provider.PutClickUpTaskResponse
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

	response, err = a.client.CreateTask(request)
	if err != nil {
		return errors.Wrap(err, "Can't create a task in ClickUp")
	}

	payload.ClickupID = response.ID
	payload.ClickupUrl = response.Url
	err = a.publisher.ClickUpTaskCreated(payload)
	if err != nil {
		return err
	}

	return nil
}

func (a *TaskCreateClickupAction) generateTaskRequest(payload model.TaskPayload) (*provider.PutClickUpTaskRequest, error) {
	request := new(provider.PutClickUpTaskRequest)

	request.Name = payload.Title
	request.NotifyAll = false
	request.Status = "To Do"
	request.Description = payload.Description + "\n" + payload.AC
	request.AddCustomField(provider.RequestedBy, payload.SlackReporter)
	request.AddCustomField(provider.SlackLink, payload.Details["slack"])

	return request, nil
}
