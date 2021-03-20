package jira

import (
	"fmt"
	"strings"
	"x-qdo/jiraclick/pkg/model"

	"github.com/andygrunwald/go-jira"
)

type ConnectorPool struct {
	clients map[string]ClientInterface
}

func NewJiraConnector(accounts map[string]model.JiraAccount) (*ConnectorPool, error) {
	clients := make(map[string]ClientInterface)

	for tenant, account := range accounts {
		tenant = strings.ToLower(tenant)
		tp := jira.BasicAuthTransport{
			Username: account.Username,
			Password: account.APIToken,
		}

		client, err := jira.NewClient(tp.Client(), account.BaseURL)
		if err != nil {
			return nil, err
		}

		clients[tenant] = &jiraClient{
			client:  client,
			project: account.Project,
			baseURL: account.BaseURL,
		}
	}

	return &ConnectorPool{
		clients: clients,
	}, nil
}

func (pool *ConnectorPool) GetInstance(tenant string) ClientInterface {
	tenant = strings.ToLower(tenant)
	if _, ok := pool.clients[tenant]; !ok {
		panic(fmt.Sprintf("tenant %s must be declared in config.yaml file", tenant))
	}
	return pool.clients[tenant]
}
