package consumer

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/trivago/tgo/tcontainer"
	"go.opentelemetry.io/otel"

	"x-qdo/jiraclick/pkg/contract"
	"x-qdo/jiraclick/pkg/model"
	"x-qdo/jiraclick/pkg/provider/jira"
	"x-qdo/jiraclick/pkg/publisher"
)

type TaskCreateJiraAction struct {
	client    *jira.ConnectorPool
	publisher *publisher.EventPublisher
}

func NewTaskCreateJiraAction(jira *jira.ConnectorPool, p *publisher.EventPublisher) (contract.Action, error) {
	return &TaskCreateJiraAction{
		client:    jira,
		publisher: p,
	}, nil
}

func (a *TaskCreateJiraAction) ProcessAction(ctx context.Context, delivery amqp.Delivery) error {
	var (
		input   inputBody
		payload model.TaskPayload
	)

	ctx, span := otel.Tracer("jira action").Start(ctx, "ProcessAction")
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

	task := a.generateTaskRequest(payload)
	span.AddEvent("Request payload generated")

	response, err := a.client.GetInstance(payload.SlackChannel).CreateIssue(ctx, task)
	if err != nil {
		err = errors.Wrap(err, "Can't create task in Jira")
		span.RecordError(err)
		return err
	}

	span.AddEvent("issue created")

	payload.JiraID = response.ID
	payload.Details["jira_url"] = response.URL
	err = a.publisher.JiraTaskCreated(ctx, payload)
	if err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("result sent to BRP")

	return nil
}

func (a *TaskCreateJiraAction) generateTaskRequest(payload model.TaskPayload) *jira.Task {
	task := new(jira.Task)

	task.Title = payload.Title
	task.Reporter = payload.GetReporterEmail()
	task.Type = "Story"
	task.Description = payload.Description + "\n\n" + payload.AC

	customFields := tcontainer.NewMarshalMap()
	customFields["customfield_10101"] = map[string]string{"value": "Internal"}
	customFields["customfield_13400"] = map[string]string{"value": "No"}
	customFields["customfield_15117"] = map[string]string{"value": "Team DevOps"}
	task.CustomFields = customFields

	return task
}
