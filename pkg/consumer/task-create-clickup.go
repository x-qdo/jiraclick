package consumer

import (
	"encoding/json"
	"github.com/araddon/dateparse"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"

	"x-qdo/jiraclick/pkg/contract"
	"x-qdo/jiraclick/pkg/model"
	"x-qdo/jiraclick/pkg/provider/clickup"
	"x-qdo/jiraclick/pkg/publisher"
)

type TaskCreateClickupAction struct {
	client    *clickup.ConnectorPool
	publisher *publisher.EventPublisher
}

func NewTaskCreateClickupAction(clickup *clickup.ConnectorPool, p *publisher.EventPublisher) (contract.Action, error) {
	return &TaskCreateClickupAction{
		client:    clickup,
		publisher: p,
	}, nil
}

func (a *TaskCreateClickupAction) ProcessAction(delivery amqp.Delivery) error {
	var (
		input   inputBody
		task    *clickup.Task
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

	request := a.generateTaskRequest(&payload)
	task, err = a.client.GetInstance(payload.SlackChannel).CreateTask(request)
	if err != nil {
		return errors.Wrap(err, "Can't create a task in ClickUp")
	}

	payload.ClickupID = task.ID
	payload.Details["clickup_url"] = task.URL
	err = a.publisher.ClickUpTaskCreated(payload)
	if err != nil {
		return err
	}

	return nil
}

func (a *TaskCreateClickupAction) generateTaskRequest(payload *model.TaskPayload) *clickup.PutClickUpTaskRequest {
	request := new(clickup.PutClickUpTaskRequest)

	request.Name = payload.Title
	request.NotifyAll = false
	request.Status = a.client.GetInstance(payload.SlackChannel).GetInitialTaskStatus()
	request.Description = payload.Description + "\n" + payload.AC
	request.AddCustomField(clickup.RequestedBy, payload.SlackReporter)
	request.AddCustomField(clickup.SlackLink, payload.Details["slack"])
	request.AddCustomField(clickup.Synced, false)

	if payload.Type == model.IncidentTaskType {
		request.Name = "[IN] " + request.Name
		payload.Title = request.Name
		request.Tags = make([]string, 0)
		request.Tags = append(request.Tags, string(payload.Type))
	}

	if payload.DueDate != "" {
		if time, err := dateparse.ParseAny(payload.DueDate); err == nil {
			timestamp := time.UnixNano() / 1e6
			request.DueDate = &timestamp
		}
	}

	return request
}
