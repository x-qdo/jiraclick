package contract

import (
	"context"
	"x-qdo/jiraclick/pkg/model"
)

type Storage interface {
	Close() error
	Begin(ctx context.Context) error
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error

	GetJiraAccounts(ctx context.Context) (map[string]model.JiraAccount, error)
	GetClickUpAccounts(ctx context.Context) (map[string]model.ClickUpAccount, error)
}
