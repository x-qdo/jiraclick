package contract

const (
	BRPActionsExchange = "actions"
)

type RoutingKey string

const (
	TaskCreate RoutingKey = "message:new.task"
)
