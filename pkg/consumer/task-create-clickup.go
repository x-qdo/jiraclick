package consumer

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"x-qdo/jiraclick/pkg/config"
	"x-qdo/jiraclick/pkg/contract"
	"x-qdo/jiraclick/pkg/provider"
	"x-qdo/jiraclick/pkg/publisher"
)

type inputBody struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Data        struct {
		Payload string `json:"payload"`
	} `json:"data"`
}

type payload struct {
	ID             string            `json:"id"`
	Title          string            `json:"title"`
	Description    string            `json:"description"`
	Details        map[string]string `json:"details"`
	SlackChannel   string            `json:"slackChannel"`
	SlackReporter  string            `json:"slackReporter"`
	SlackTS        string            `json:"slackTS"`
	LastUpdateTime string            `json:"LastUpdateTime"`
	DueDate        string            `json:"dueDate"`
	AC             string            `json:"ac"`
}

type TaskCreateClickupAction struct {
	client    *provider.ClickUpAPIClient
	publisher *publisher.EventPublisher
}

func NewTaskCreateClickupAction(cfg *config.Config, clickup *provider.ClickUpAPIClient, p *publisher.EventPublisher) (contract.Action, error) {
	return &TaskCreateClickupAction{
		client:    clickup,
		publisher: p,
	}, nil
}

func (a *TaskCreateClickupAction) ProcessAction(delivery amqp.Delivery) error {

	var (
		inputBody inputBody
		response  *provider.PutTaskResponse
		payload   payload
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

	response.SlackBotID = payload.ID
	err = a.publisher.ClickUpTaskCreated(response, payload.SlackChannel)
	if err != nil {
		return err
	}

	return nil
}

func (a *TaskCreateClickupAction) generateTaskRequest(payload payload) (*provider.PutTaskRequest, error) {
	request := new(provider.PutTaskRequest)

	request.Name = payload.Title
	request.NotifyAll = false
	request.Status = "To Do"
	request.Description = payload.Description + "\n" + payload.AC
	request.AddCustomField(provider.RequestedBy, payload.SlackReporter)
	request.AddCustomField(provider.SlackLink, payload.Details["slack"])

	return request, nil
}
