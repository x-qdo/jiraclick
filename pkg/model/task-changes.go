package model

type TaskChanges struct {
	Type      string   `json:"type"`
	ClickupID string   `json:"clickup_id,omitempty"`
	JiraID    string   `json:"jira_id,omitempty"`
	Changes   []Change `json:"changes"`
	Username  string   `json:"username"`
}

type Change struct {
	Field    string      `json:"field"`
	NewValue interface{} `json:"new_value"`
}

func (t *TaskChanges) AddChange(field string, value interface{}) {
	t.Changes = append(t.Changes, Change{
		Field:    field,
		NewValue: value,
	})
}
