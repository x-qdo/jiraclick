package consumer

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"github.com/trivago/tgo/tcontainer"
	"x-qdo/jiraclick/pkg/contract"
	"x-qdo/jiraclick/pkg/model"
	"x-qdo/jiraclick/pkg/provider"
	"x-qdo/jiraclick/pkg/publisher"
)

type TaskCreateJiraAction struct {
	client    *provider.JiraClient
	publisher *publisher.EventPublisher
}

func NewTaskCreateJiraAction(jira *provider.JiraClient, p *publisher.EventPublisher) (contract.Action, error) {
	return &TaskCreateJiraAction{
		client:    jira,
		publisher: p,
	}, nil
}

func (a *TaskCreateJiraAction) ProcessAction(delivery amqp.Delivery) error {
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

	task, err := a.generateTaskRequest(payload)
	if err != nil {
		return errors.Wrap(err, "Can't create task request")
	}

	response, err := a.client.CreateIssue(task)
	if err != nil {
		return errors.Wrap(err, "Can't create task in Jira")
	}

	payload.JiraID = response.ID
	payload.Details["jira_url"] = response.Url
	err = a.publisher.JiraTaskCreated(payload)
	if err != nil {
		return err
	}

	return nil
}

func (a *TaskCreateJiraAction) generateTaskRequest(payload model.TaskPayload) (*provider.JiraTask, error) {
	task := new(provider.JiraTask)

	task.Title = payload.Title
	task.Reporter = payload.GetReporterEmail()
	task.Type = "Story"
	task.Description = payload.Description + "\n\n" + payload.AC

	customFields := tcontainer.NewMarshalMap()
	customFields["customfield_10101"] = map[string]string{"value": "Internal"}
	customFields["customfield_13400"] = map[string]string{"value": "No"}
	customFields["customfield_15117"] = map[string]string{"value": "Team DevOps"}
	task.CustomFields = customFields

	return task, nil
}
