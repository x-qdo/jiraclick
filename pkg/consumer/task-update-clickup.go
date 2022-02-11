package consumer

import (
	"context"
	"encoding/json"
	"go.opentelemetry.io/otel"

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

func (a *TaskUpdateClickupAction) ProcessAction(ctx context.Context, delivery amqp.Delivery) error {
	var (
		input   inputBody
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

	request := a.generateTaskRequest(payload)
	span.AddEvent("Request payload generated")
	err = a.client.GetInstance(payload.SlackChannel).UpdateTask(ctx, payload.ClickupID, request)
	if err != nil {
		err = errors.Wrap(err, "Can't update a task in ClickUp")
		span.RecordError(err)
		return err
	}

	return nil
}

func (a *TaskUpdateClickupAction) generateTaskRequest(payload model.TaskPayload) *clickup.PutClickUpTaskRequest {
	request := new(clickup.PutClickUpTaskRequest)

	request.Name = payload.Title
	request.Description = payload.Description + "\n\n" + payload.AC
	request.AddCustomField(clickup.RequestedBy, payload.SlackReporter)
	request.AddCustomField(clickup.SlackLink, payload.Details["slack"])
	request.AddCustomField(clickup.JiraLink, payload.Details["clickup_url"])

	return request
}
