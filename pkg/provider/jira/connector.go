package jira

import (
	"fmt"
	"strings"

	"github.com/andygrunwald/go-jira"

	"x-qdo/jiraclick/pkg/config"
)

type ConnectorPool struct {
	clients map[string]ClientInterface
}

func NewJiraConnector(cfg *config.Config) (*ConnectorPool, error) {
	clients := make(map[string]ClientInterface)

	for tenant, instance := range cfg.Jira {
		tenant = strings.ToLower(tenant)
		tp := jira.BasicAuthTransport{
			Username: instance.Username,
			Password: instance.APIToken,
		}

		client, err := jira.NewClient(tp.Client(), instance.BaseURL)
		if err != nil {
			return nil, err
		}

		clients[tenant] = &jiraClient{
			client:  client,
			project: instance.Project,
			baseURL: instance.BaseURL,
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
