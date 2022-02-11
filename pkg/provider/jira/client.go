package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/andygrunwald/go-jira"
	"github.com/pkg/errors"
	"github.com/trivago/tgo/tcontainer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const LinkToJiraTask = "%s/browse/%s"

type ClientInterface interface {
	CreateIssue(ctx context.Context, task *Task) (*PutJiraTaskResponse, error)
	UpdateIssue(ctx context.Context) error
	FindUserByEmail(ctx context.Context, email string) *jira.User
}

type jiraClient struct {
	client  *jira.Client
	project string
	baseURL string
}

type Task struct {
	ID           string
	Title        string
	Description  string
	Reporter     string
	Type         string
	CustomFields tcontainer.MarshalMap
}

type PutJiraTaskResponse struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

func (c *jiraClient) CreateIssue(ctx context.Context, task *Task) (*PutJiraTaskResponse, error) {
	var response PutJiraTaskResponse

	ctx, span := otel.Tracer("jira client").Start(ctx, "CreateIssue")
	defer span.End()

	t, err := json.Marshal(task)
	if err != nil {
		span.RecordError(err)
	} else {
		span.SetAttributes(attribute.Key("task").String(string(t)))
	}

	i := jira.Issue{
		Fields: &jira.IssueFields{
			Reporter:    c.FindUserByEmail(ctx, task.Reporter),
			Description: task.Description,
			Type: jira.IssueType{
				Name: task.Type,
			},
			Project: jira.Project{
				Key: c.project,
			},
			Summary:  task.Title,
			Unknowns: task.CustomFields,
		},
	}

	issue, r, err := c.client.Issue.CreateWithContext(ctx, &i)
	if err != nil {
		span.RecordError(err)
		buf := new(bytes.Buffer)
		if _, e := buf.ReadFrom(r.Response.Body); e != nil {
			span.RecordError(e)
			return nil, e
		}
		return nil, errors.Wrap(err, buf.String())
	}
	task.ID = issue.ID

	response.ID = issue.ID
	response.URL = fmt.Sprintf(LinkToJiraTask, c.baseURL, issue.Key)
	span.AddEvent("issue has been created", trace.WithAttributes(
		attribute.Key("issue id").String(issue.ID),
		attribute.Key("issue url").String(response.URL),
	))

	return &response, nil
}

func (c *jiraClient) UpdateIssue(ctx context.Context) error {
	ctx, span := otel.Tracer("jira client").Start(ctx, "UpdateIssue")
	defer span.End()
	return nil
}

func (c *jiraClient) FindUserByEmail(ctx context.Context, email string) *jira.User {
	ctx, span := otel.Tracer("jira client").Start(ctx, "FindUserByEmail")
	defer span.End()
	span.SetAttributes(attribute.Key("email").String(email))

	users, _, err := c.client.User.FindWithContext(ctx, email)
	if err != nil {
		span.RecordError(err)
		return nil
	}

	usersStr := make([]string, 0)
	for _, u := range users {
		usersStr = append(usersStr, u.Name)
	}
	span.AddEvent("following users has been found", trace.WithAttributes(
		attribute.Key("users").StringSlice(usersStr),
	))

	if len(users) > 1 {
		span.AddEvent("user selected", trace.WithAttributes(
			attribute.Key("user").String(users[0].Name),
		))
		return &users[0]
	}

	return nil
}
