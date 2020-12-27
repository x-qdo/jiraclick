package clickup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"x-qdo/jiraclick/pkg/config"
)

type APIClient struct {
	httpClient http.Client
	options    struct {
		host   string
		token  string
		listID string
	}
}

type PutClickUpTaskRequest struct {
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Status       string        `json:"status,omitempty"`
	NotifyAll    bool          `json:"notify_all,omitempty"`
	CustomFields []CustomField `json:"custom_fields,omitempty"`
	Tags         []string      `json:"tags"`
}

func (t *PutClickUpTaskRequest) AddCustomField(id CustomFieldKey, value interface{}) {
	t.CustomFields = append(t.CustomFields, CustomField{ID: id, Value: value})
}

func NewClickUpClient(cfg *config.Config) (*APIClient, error) {
	client := new(APIClient)
	client.options.host = cfg.ClickUp.Host
	client.options.token = cfg.ClickUp.Token
	client.options.listID = cfg.ClickUp.List

	return client, nil
}

func (c *APIClient) CreateTask(request *PutClickUpTaskRequest) (*Task, error) {
	var task Task
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", c.options.host+"/list/"+c.options.listID+"/task/", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", c.options.token)
	req.Header.Add("Content-Type", "application/json")

	r, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ClickUp API error: %s", r.Status)
	}
	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (c *APIClient) UpdateTask(taskID string, request *PutClickUpTaskRequest) error {
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", c.options.host+"/task/"+taskID, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", c.options.token)
	req.Header.Add("Content-Type", "application/json")

	r, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("ClickUp API error: %s", r.Status)
	}
	defer r.Body.Close()

	var customField CustomField
	for _, customField = range request.CustomFields {
		err = c.SetCustomField(taskID, string(customField.ID), customField.Value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *APIClient) SetCustomField(taskID, customFieldID string, value interface{}) error {
	var request struct {
		Value interface{} `json:"value"`
	}
	request.Value = value
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", c.options.host+"/task/"+taskID+"/field/"+customFieldID, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", c.options.token)
	req.Header.Add("Content-Type", "application/json")

	r, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("ClickUp API error: %s", r.Status)
	}
	defer r.Body.Close()

	return nil
}

func (c *APIClient) GetTask(taskID string) (*Task, error) {
	var task Task
	req, err := http.NewRequest("GET", c.options.host+"/task/"+taskID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", c.options.token)
	req.Header.Add("Content-Type", "application/json")

	r, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ClickUp API error: %s", r.Status)
	}
	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}
