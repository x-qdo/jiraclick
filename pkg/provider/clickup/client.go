package clickup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
)

type APIClient struct {
	httpClient http.Client
	options    struct {
		host              string
		token             string
		listID            string
		initialTaskStatus string
	}
}

type ClientInterface interface {
	CreateTask(request *PutClickUpTaskRequest) (*Task, error)
	UpdateTask(taskID string, request *PutClickUpTaskRequest) error
	SetCustomField(taskID, customFieldID string, value interface{}) error
	GetTask(taskID string) (*Task, error)
	GetInitialTaskStatus() string
}

type PutClickUpTaskRequest struct {
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Status       string        `json:"status,omitempty"`
	NotifyAll    bool          `json:"notify_all,omitempty"`
	CustomFields []CustomField `json:"custom_fields,omitempty"`
	Tags         []string      `json:"tags"`
	DueDate      *int64        `json:"due_date,omitempty"`
}

func (t *PutClickUpTaskRequest) AddCustomField(id CustomFieldKey, value interface{}) {
	t.CustomFields = append(t.CustomFields, CustomField{ID: id, Value: value})
}

func (c *APIClient) CreateTask(request *PutClickUpTaskRequest) (*Task, error) {
	var task Task
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	logrus.Debug("Sending POST request to Clickup: ", c.options.host+"/list/"+c.options.listID+"/task/")
	logrus.Debug(string(body))
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
		return nil, formatHttpError(r)
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
	logrus.Debug("Sending PUT request to Clickup: ", c.options.host+"/task/"+taskID)
	logrus.Debug(string(body))
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
		return formatHttpError(r)
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

	logrus.Debug("Sending POST request to Clickup: ", c.options.host+"/task/"+taskID+"/field/"+customFieldID)
	logrus.Debug(string(body))
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
		return formatHttpError(r)
	}
	defer r.Body.Close()

	return nil
}

func (c *APIClient) GetTask(taskID string) (*Task, error) {
	var task Task
	logrus.Debug("Sending GET request to Clickup: ", c.options.host+"/task/"+taskID)
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
		return nil, formatHttpError(r)
	}
	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (c *APIClient) GetInitialTaskStatus() string {
	return c.options.initialTaskStatus
}

func formatHttpError(r *http.Response) error {
	dump, _ := httputil.DumpResponse(r, true)
	return fmt.Errorf("ClickUp API error status: %s body: %q", r.Status, dump)
}
