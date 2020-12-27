package jira

import (
	"bytes"
	"fmt"
	"github.com/andygrunwald/go-jira"
	"github.com/pkg/errors"
	"github.com/trivago/tgo/tcontainer"
)

const LinkToJiraTask = "%s/browse/%s"

type ClientInterface interface {
	CreateIssue(task *JiraTask) (*PutJiraTaskResponse, error)
	UpdateIssue() error
	FindUserByEmail(email string) *jira.User
}

type jiraClient struct {
	client  *jira.Client
	project string
	baseURL string
}

type JiraTask struct {
	ID           string
	Title        string
	Description  string
	Reporter     string
	Type         string
	CustomFields tcontainer.MarshalMap
}

type PutJiraTaskResponse struct {
	ID  string `json:"id"`
	Url string `json:"url"`
}

func (c *jiraClient) CreateIssue(task *JiraTask) (*PutJiraTaskResponse, error) {
	var response PutJiraTaskResponse

	i := jira.Issue{
		Fields: &jira.IssueFields{
			Reporter:    c.FindUserByEmail(task.Reporter),
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

	issue, r, err := c.client.Issue.Create(&i)
	if err != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Response.Body)
		return nil, errors.Wrap(err, buf.String())
	}
	task.ID = issue.ID

	response.ID = issue.ID
	response.Url = fmt.Sprintf(LinkToJiraTask, c.baseURL, issue.Key)

	return &response, nil
}

func (c *jiraClient) UpdateIssue() error {
	return nil
}

func (c *jiraClient) FindUserByEmail(email string) *jira.User {
	users, _, err := c.client.User.Find(email)
	if err != nil {
		return nil
	} else if len(users) > 1 {
		return &users[0]
	}

	return nil
}
