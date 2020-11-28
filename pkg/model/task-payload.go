package model

import "strings"

type TaskPayload struct {
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
	ClickupID      string            `json:"clickup_id"`
	ClickupUrl     string            `json:"clickup_url"`
	JiraID         string            `json:"jira_id"`
	JiraUrl        string            `json:"jira_url"`
}

func (p *TaskPayload) GetReporterEmail() string {
	if reporterDetails := strings.Split(p.Details["reporter"], "||"); len(reporterDetails) >= 3 {
		return reporterDetails[1]
	}

	return ""
}
