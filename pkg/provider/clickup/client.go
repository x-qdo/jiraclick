package clickup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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
	CreateTask(ctx context.Context, request *PutClickUpTaskRequest) (*Task, error)
	UpdateTask(ctx context.Context, taskID string, request *PutClickUpTaskRequest) error
	SetCustomField(ctx context.Context, taskID, customFieldID string, value interface{}) error
	GetTask(ctx context.Context, taskID string) (*Task, error)
	GetInitialTaskStatus(ctx context.Context) string
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

func (c *APIClient) CreateTask(ctx context.Context, request *PutClickUpTaskRequest) (*Task, error) {
	var task Task
	ctx, span := otel.Tracer("clickup provider").Start(ctx, "CreateTask")
	defer span.End()

	body, err := json.Marshal(request)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	span.SetAttributes(
		attribute.String("url", c.options.host+"/list/"+c.options.listID+"/task/"),
		attribute.String("request body", string(body)),
	)
	req, err := http.NewRequest("POST", c.options.host+"/list/"+c.options.listID+"/task/", bytes.NewBuffer(body))
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	req.Header.Add("Authorization", c.options.token)
	req.Header.Add("Content-Type", "application/json")

	r, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	span.AddEvent("POST request sent to ClickUp")

	if r.StatusCode != http.StatusOK {
		err = formatHttpError(r)
		span.RecordError(err)
		return nil, err
	}
	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return &task, nil
}

func (c *APIClient) UpdateTask(ctx context.Context, taskID string, request *PutClickUpTaskRequest) error {
	ctx, span := otel.Tracer("clickup provider").Start(ctx, "UpdateTask")
	defer span.End()

	body, err := json.Marshal(request)
	if err != nil {
		span.RecordError(err)
		return err
	}
	span.SetAttributes(
		attribute.String("url", c.options.host+"/task/"+taskID),
		attribute.String("request body", string(body)),
	)
	req, err := http.NewRequest("PUT", c.options.host+"/task/"+taskID, bytes.NewBuffer(body))
	if err != nil {
		span.RecordError(err)
		return err
	}
	req.Header.Add("Authorization", c.options.token)
	req.Header.Add("Content-Type", "application/json")

	r, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("POST request sent to ClickUp")

	if r.StatusCode != http.StatusOK {
		err = formatHttpError(r)
		span.RecordError(err)
		return err
	}
	defer r.Body.Close()

	var customField CustomField
	for _, customField = range request.CustomFields {
		err = c.SetCustomField(ctx, taskID, string(customField.ID), customField.Value)
		if err != nil {
			span.RecordError(err)
			return err
		}
	}

	return nil
}

func (c *APIClient) SetCustomField(ctx context.Context, taskID, customFieldID string, value interface{}) error {
	var request struct {
		Value interface{} `json:"value"`
	}

	ctx, span := otel.Tracer("clickup provider").Start(ctx, "SetCustomField")
	defer span.End()

	request.Value = value
	body, err := json.Marshal(request)
	if err != nil {
		span.RecordError(err)
		return err
	}

	span.SetAttributes(
		attribute.String("url", c.options.host+"/task/"+taskID+"/field/"+customFieldID),
		attribute.String("request body", string(body)),
	)
	req, err := http.NewRequest("POST", c.options.host+"/task/"+taskID+"/field/"+customFieldID, bytes.NewBuffer(body))
	if err != nil {
		span.RecordError(err)
		return err
	}
	req.Header.Add("Authorization", c.options.token)
	req.Header.Add("Content-Type", "application/json")

	r, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("POST request sent to ClickUp")

	if r.StatusCode != http.StatusOK {
		err = formatHttpError(r)
		span.RecordError(err)
		return err
	}
	defer r.Body.Close()

	return nil
}

func (c *APIClient) GetTask(ctx context.Context, taskID string) (*Task, error) {
	var task Task
	ctx, span := otel.Tracer("clickup provider").Start(ctx, "GetTask")
	defer span.End()
	span.SetAttributes(
		attribute.String("url", c.options.host+"/task/"+taskID),
	)

	req, err := http.NewRequest("GET", c.options.host+"/task/"+taskID, nil)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	req.Header.Add("Authorization", c.options.token)
	req.Header.Add("Content-Type", "application/json")

	r, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.AddEvent("GET request sent to ClickUp")

	if r.StatusCode != http.StatusOK {
		return nil, formatHttpError(r)
	}
	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.AddEvent("task received")

	return &task, nil
}

func (c *APIClient) GetInitialTaskStatus(ctx context.Context) string {
	ctx, span := otel.Tracer("clickup provider").Start(ctx, "GetInitialTaskStatus")
	defer span.End()
	return c.options.initialTaskStatus
}

func formatHttpError(r *http.Response) error {
	dump, _ := httputil.DumpResponse(r, true)
	return fmt.Errorf("ClickUp API error status: %s body: %q", r.Status, dump)
}
