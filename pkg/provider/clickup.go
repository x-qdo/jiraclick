package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"x-qdo/jiraclick/pkg/config"
)

type CustomFieldKey string

const (
	ApprovedBy       CustomFieldKey = "065a1567-0655-4a7e-aefe-e179f7983069"
	BillableHours    CustomFieldKey = "074e1387-e7b8-41c6-92db-fbada8f8486c"
	JiraLink         CustomFieldKey = "349fbec4-f71f-4cee-9861-c112e253a6e1"
	SlackLink        CustomFieldKey = "517a450f-ce8b-4683-b34a-616d5c3b0fb4"
	DoneNotification CustomFieldKey = "86477c9c-b494-423b-8dc0-3a49734b8b28"
	Synced           CustomFieldKey = "926d35a9-5f70-4f54-bc07-d11b82d4cf21"
	RequestedBy      CustomFieldKey = "eb30f61c-dbad-4ad4-896d-15d2a239cb69"
)

type ClickUpAPIClient struct {
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
	CustomFields []customField `json:"custom_fields,omitempty"`
}

type PutClickUpTaskResponse struct {
	ID  string `json:"id"`
	Url string `json:"url"`
}

type customField struct {
	ID    CustomFieldKey `json:"id"`
	Value interface{}    `json:"value"`
}

func (t *PutClickUpTaskRequest) AddCustomField(id CustomFieldKey, value interface{}) {
	t.CustomFields = append(t.CustomFields, customField{ID: id, Value: value})
}

func NewClickUpClient(cfg *config.Config) (*ClickUpAPIClient, error) {
	client := new(ClickUpAPIClient)
	client.options.host = cfg.ClickUp.Host
	client.options.token = cfg.ClickUp.Token
	client.options.listID = cfg.ClickUp.List

	return client, nil
}

func (c *ClickUpAPIClient) CreateTask(request *PutClickUpTaskRequest) (*PutClickUpTaskResponse, error) {
	var response PutClickUpTaskResponse
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest("POST", c.options.host+"/list/"+c.options.listID+"/task/", bytes.NewBuffer(body))
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
	err = json.NewDecoder(r.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *ClickUpAPIClient) UpdateTask(taskID string, request *PutClickUpTaskRequest) error {
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}
	req, _ := http.NewRequest("PUT", c.options.host+"/task/"+taskID, bytes.NewBuffer(body))
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

	var customField customField
	for _, customField = range request.CustomFields {
		err = c.SetCustomField(taskID, string(customField.ID), customField.Value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *ClickUpAPIClient) SetCustomField(taskID, customFieldID string, value interface{}) error {
	var request struct {
		Value interface{} `json:"value"`
	}
	request.Value = value
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}
	req, _ := http.NewRequest("POST", c.options.host+"/task/"+taskID+"/field/"+customFieldID, bytes.NewBuffer(body))
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
