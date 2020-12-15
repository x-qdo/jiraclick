package consumer

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"x-qdo/jiraclick/pkg/contract"
	"x-qdo/jiraclick/pkg/model"
	"x-qdo/jiraclick/pkg/provider/clickup"
	"x-qdo/jiraclick/pkg/publisher"
)

type TaskCreateClickupAction struct {
	client    *clickup.ClickUpAPIClient
	publisher *publisher.EventPublisher
}

func NewTaskCreateClickupAction(clickup *clickup.ClickUpAPIClient, p *publisher.EventPublisher) (contract.Action, error) {
	return &TaskCreateClickupAction{
		client:    clickup,
		publisher: p,
	}, nil
}

func (a *TaskCreateClickupAction) ProcessAction(delivery amqp.Delivery) error {

	var (
		inputBody inputBody
		task      *clickup.Task
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

	request, err := a.generateTaskRequest(&payload)
	if err != nil {
		return errors.Wrap(err, "Can't create task request")
	}

	task, err = a.client.CreateTask(request)
	if err != nil {
		return errors.Wrap(err, "Can't create a task in ClickUp")
	}

	payload.ClickupID = task.ID
	payload.Details["clickup_url"] = task.Url
	err = a.publisher.ClickUpTaskCreated(payload)
	if err != nil {
		return err
	}

	return nil
}

func (a *TaskCreateClickupAction) generateTaskRequest(payload *model.TaskPayload) (*clickup.PutClickUpTaskRequest, error) {
	request := new(clickup.PutClickUpTaskRequest)

	request.Name = payload.Title
	request.NotifyAll = false
	request.Status = "To Do"
	request.Description = payload.Description + "\n" + payload.AC
	request.AddCustomField(clickup.RequestedBy, payload.SlackReporter)
	request.AddCustomField(clickup.SlackLink, payload.Details["slack"])
	request.AddCustomField(clickup.Synced, true)

	if payload.Type == model.IncidentTaskType {
		request.Name = "[IN] " + request.Name
		payload.Title = request.Name
		request.Tags = make([]string, 0)
		request.Tags = append(request.Tags, string(payload.Type))
	}

	return request, nil
}
