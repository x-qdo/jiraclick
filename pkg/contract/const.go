package contract

const (
	BRPActionsExchange = "actions"
	BRPEventsExchange  = "events"
)

type RoutingKey string

const (
	TaskCreateClickUp RoutingKey = "task:create.clickup"
	TaskCreateJira    RoutingKey = "task:create.jira"

	TaskCreatedClickUpEvent RoutingKey = "t:%s:clickup:task.created"
	TaskCreatedJiraEvent    RoutingKey = "t:%s:jira:task.created"
)
