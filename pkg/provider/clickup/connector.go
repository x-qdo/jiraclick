package clickup

import (
	"fmt"
	"strings"
	"x-qdo/jiraclick/pkg/model"
)

type ConnectorPool struct {
	clients map[string]ClientInterface
}

func NewClickUpConnector(accounts map[string]model.ClickUpAccount) (*ConnectorPool, error) {
	clients := make(map[string]ClientInterface)

	for tenant, account := range accounts {
		tenant = strings.ToLower(tenant)
		client := new(APIClient)
		client.options.host = account.Host
		client.options.token = account.Token
		client.options.listID = account.List
		client.options.initialTaskStatus = account.InitialTaskStatus

		clients[tenant] = client
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
