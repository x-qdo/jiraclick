package consumer

import (
	"encoding/json"

	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"

	"x-qdo/jiraclick/pkg/contract"
	"x-qdo/jiraclick/pkg/model"
	"x-qdo/jiraclick/pkg/provider/clickup"
	"x-qdo/jiraclick/pkg/publisher"
)

type TaskUpdateClickupAction struct {
	client    *clickup.ConnectorPool
	publisher *publisher.EventPublisher
}

func NewTaskUpdateClickupAction(clickup *clickup.ConnectorPool, p *publisher.EventPublisher) (contract.Action, error) {
	return &TaskUpdateClickupAction{
		client:    clickup,
		publisher: p,
	}, nil
}

func (a *TaskUpdateClickupAction) ProcessAction(delivery amqp.Delivery) error {
	var (
		input   inputBody
		payload model.TaskPayload
	)

	err := json.Unmarshal(delivery.Body, &input)
	if err != nil {
		return errors.Wrap(err, "Can't unmarshall task body")
	}

	err = json.Unmarshal([]byte(input.Data.Payload), &payload)
	if err != nil {
		return errors.Wrap(err, "Can't unmarshall task body")
	}

	request, err := a.generateTaskRequest(payload)
	if err != nil {
		return errors.Wrap(err, "Can't create task request")
	}

	err = a.client.GetInstance(payload.SlackChannel).UpdateTask(payload.ClickupID, request)
	if err != nil {
		return errors.Wrap(err, "Can't update a task in ClickUp")
	}

	return nil
}

func (a *TaskUpdateClickupAction) generateTaskRequest(payload model.TaskPayload) (*clickup.PutClickUpTaskRequest, error) {
	request := new(clickup.PutClickUpTaskRequest)

	request.Name = payload.Title
	request.Description = payload.Description + "\n\n" + payload.AC
	request.AddCustomField(clickup.RequestedBy, payload.SlackReporter)
	request.AddCustomField(clickup.SlackLink, payload.Details["slack"])
	request.AddCustomField(clickup.JiraLink, payload.Details["clickup_url"])

	return request, nil
}
