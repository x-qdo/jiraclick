package consumer

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"github.com/trivago/tgo/tcontainer"

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

func (a *TaskCreateJiraAction) ProcessAction(delivery amqp.Delivery) error {
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

	task := a.generateTaskRequest(payload)

	response, err := a.client.GetInstance(payload.SlackChannel).CreateIssue(task)
	if err != nil {
		return errors.Wrap(err, "Can't create task in Jira")
	}

	payload.JiraID = response.ID
	payload.Details["jira_url"] = response.URL
	err = a.publisher.JiraTaskCreated(payload)
	if err != nil {
		return err
	}

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
