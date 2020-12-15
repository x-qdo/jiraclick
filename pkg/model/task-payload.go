package model

import "strings"

type taskType string

const (
	RegularTaskType  taskType = "regular"
	IncidentTaskType taskType = "incident"
)

type TaskPayload struct {
	ID             string            `json:"id"`
	Type           taskType          `json:"type"`
	Title          string            `json:"title"`
	Description    string            `json:"description"`
	Details        map[string]string `json:"details"`
	SlackChannel   string            `json:"slackChannel"`
	SlackReporter  string            `json:"slackReporter"`
	SlackTS        string            `json:"slackTS"`
	LastUpdateTime string            `json:"LastUpdateTime"`
	DueDate        string            `json:"dueDate"`
	AC             string            `json:"ac"`
	ClickupID      string            `json:"clickup_id"`
	JiraID         string            `json:"jira_id"`
}

func (p *TaskPayload) GetReporterEmail() string {
	if reporterDetails := strings.Split(p.Details["reporter"], "||"); len(reporterDetails) >= 3 {
		return reporterDetails[1]
	}

	return ""
}
