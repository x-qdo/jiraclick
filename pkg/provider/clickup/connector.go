package clickup

import (
	"fmt"
	"strings"
	"x-qdo/jiraclick/pkg/config"
)

type ConnectorPool struct {
	clients map[string]ClientInterface
}

func NewClickUpConnector(cfg *config.Config) (*ConnectorPool, error) {
	clients := make(map[string]ClientInterface)

	for tenant, instance := range cfg.ClickUp {
		tenant = strings.ToLower(tenant)
		client := new(APIClient)
		client.options.host = instance.Host
		client.options.token = instance.Token
		client.options.listID = instance.List

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
