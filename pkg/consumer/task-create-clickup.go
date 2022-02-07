package consumer

import (
	"context"
	"encoding/json"
	"github.com/araddon/dateparse"
	"go.opentelemetry.io/otel"

	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"

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

func (a *TaskCreateClickupAction) ProcessAction(ctx context.Context, delivery amqp.Delivery) error {
	var (
		input   inputBody
		task    *clickup.Task
		payload model.TaskPayload
	)

	ctx, span := otel.Tracer("clickup action").Start(ctx, "ProcessAction")
	defer span.End()

	err := json.Unmarshal(delivery.Body, &input)
	if err != nil {
		err = errors.Wrap(err, "Can't unmarshall task body")
		span.RecordError(err)
		return err
	}

	err = json.Unmarshal([]byte(input.Data.Payload), &payload)
	if err != nil {
		err = errors.Wrap(err, "Can't unmarshall task body")
		span.RecordError(err)
		return err
	}

	request := a.generateTaskRequest(ctx, &payload)
	span.AddEvent("Request payload generated")
	task, err = a.client.GetInstance(payload.SlackChannel).CreateTask(ctx, request)
	if err != nil {
		err = errors.Wrap(err, "Can't create a task in ClickUp")
		span.RecordError(err)
		return err
	}

	span.AddEvent("task created")

	payload.ClickupID = task.ID
	payload.Details["clickup_url"] = task.URL
	err = a.publisher.ClickUpTaskCreated(ctx, payload)
	if err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("result sent to BRP")

	return nil
}

func (a *TaskCreateClickupAction) generateTaskRequest(ctx context.Context, payload *model.TaskPayload) *clickup.PutClickUpTaskRequest {
	request := new(clickup.PutClickUpTaskRequest)

	ctx, span := otel.Tracer("clickup action").Start(ctx, "generateTaskRequest")
	defer span.End()

	request.Name = payload.Title
	request.NotifyAll = false
	request.Status = a.client.GetInstance(payload.SlackChannel).GetInitialTaskStatus(ctx)
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
