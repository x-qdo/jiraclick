package model

type Account struct {
	tableName    struct{}    `pg:"accounts"`
	Id           string      `pg:"type:serial"`
	SlackChannel string      `pg:"slack_channel"`
	Resource     string      `pg:"resource"`
	Props        interface{} `pg:"props,type:jsonb"`
}

type ClickUpAccount struct {
	Host              string `json:"host"`
	Token             string `json:"token"`
	List              string `json:"list"`
	WebhookSecret     string `json:"webhooksecret"`
	InitialTaskStatus string `json:"initial_status"`
}

type JiraAccount struct {
	Username string `json:"username"`
	APIToken string `json:"apitoken"`
	BaseURL  string `json:"baseurl"`
	Project  string `json:"project"`
}
