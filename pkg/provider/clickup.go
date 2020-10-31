package provider

import (
	"bytes"
	"encoding/json"
	"net/http"
	"x-qdo/jiraclick/pkg/config"
)

type ClickUpAPIClient struct {
	httpClient http.Client
	options    struct {
		token  string
		listID string
	}
}

type PutTaskRequest struct {
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Status       string        `json:"status"`
	NotifyAll    bool          `json:"notify_all"`
	CustomFields []customField `json:"custom_fields"`
}

type customField struct {
	ID    string      `json:"id"`
	Value interface{} `json:"value"`
}

func NewClickUpClient(cfg config.Config) (*ClickUpAPIClient, error) {
	client := new(ClickUpAPIClient)

	client.options.token = cfg.ClickUp.Token

	return client, nil
}

func (c *ClickUpAPIClient) CreateTask(request *PutTaskRequest) error {
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}
	req, _ := http.NewRequest("POST", "https://api.clickup.com/api/v2/list/"+c.options.listID+"/task/", bytes.NewBuffer(body))
	req.Header.Add("Authorization", c.options.token)
	req.Header.Add("Content-Type", "application/json")

	_, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}
	return nil
}
