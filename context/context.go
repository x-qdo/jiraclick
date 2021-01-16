package context

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"x-qdo/jiraclick/cmd"
	"x-qdo/jiraclick/pkg/config"
	"x-qdo/jiraclick/pkg/provider"
	"x-qdo/jiraclick/pkg/provider/clickup"
	"x-qdo/jiraclick/pkg/provider/jira"
)

type Context struct {
	Ctx       context.Context
	CancelF   context.CancelFunc
	Config    config.Config
	RootCmd   *cobra.Command
	WaitGroup *sync.WaitGroup
}

func NewContext(configPath string) (*Context, error) {
	var ctx Context
	ctx.Ctx, ctx.CancelF = context.WithCancel(context.Background())

	cfg, err := config.NewConfig(configPath)
	if err != nil {
		panic(err)
	}

	logger := logrus.New()
	if cfg.Debug {
		logger.SetLevel(logrus.DebugLevel)
		logrus.SetLevel(logrus.DebugLevel)
	}

	ctx.WaitGroup = new(sync.WaitGroup)

	amqpProvider, err := provider.NewRabbitChannel(ctx.Ctx, ctx.WaitGroup, cfg, logger)
	if err != nil {
		panic(err)
	}

	clickupProvider, err := clickup.NewClickUpConnector(cfg)
	if err != nil {
		panic(err)
	}

	jiraProvider, err := jira.NewJiraConnector(cfg)
	if err != nil {
		panic(err)
	}

	setCommands(&ctx, cfg, amqpProvider, clickupProvider, jiraProvider, logger)

	return &ctx, nil
}

func setCommands(
	ctx *Context,
	cfg *config.Config,
	queue *provider.RabbitChannel,
	clickup *clickup.ConnectorPool,
	jira *jira.ConnectorPool,
	logger *logrus.Logger,
) {
	workerCmd := cmd.NewWorkerCmd(queue, clickup, jira)
	httpHandlerCmd := cmd.NewHTTPHandlerCmd(cfg, logger, queue, clickup)

	rootCmd := cmd.NewRootCmd()

	rootCmd.AddCommand(workerCmd)
	rootCmd.AddCommand(httpHandlerCmd)

	ctx.RootCmd = rootCmd
}

func (c *Context) Done() <-chan struct{} {
	return c.Ctx.Done()
}
