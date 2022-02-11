package context

import (
	"context"
	"github.com/astreter/amqpwrapper/v2"
	"github.com/x-qdo/otelwrapper"
	"sync"
	"x-qdo/jiraclick/pkg/contract"

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

func NewContext() (*Context, error) {
	var ctx Context
	ctx.Ctx, ctx.CancelF = context.WithCancel(context.Background())

	cfg, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	logger := logrus.New()
	if cfg.Debug {
		logger.SetLevel(logrus.DebugLevel)
		logrus.SetLevel(logrus.DebugLevel)
	}

	ctx.WaitGroup = new(sync.WaitGroup)

	tp, err := otelwrapper.InitTracerProvider(config.ServiceName, "default")
	if err != nil {
		return nil, err
	}
	go otelwrapper.ShutdownWaiting(tp, ctx.Ctx, ctx.WaitGroup)

	amqpProvider, err := amqpwrapper.NewRabbitChannel(ctx.Ctx, ctx.WaitGroup, &amqpwrapper.Config{
		URL:          cfg.RabbitMQ.URL,
		Debug:        cfg.Debug,
		ConfirmSends: true,
	})
	if err != nil {
		panic(err)
	}

	go func() {
		<-amqpProvider.Cancel()
		ctx.CancelF()
	}()

	db, err := provider.NewPostgres(cfg)
	if err != nil {
		panic(err)
	}

	clickUpAccounts, err := db.GetClickUpAccounts(ctx.Ctx)
	if err != nil {
		panic(err)
	}
	clickupProvider, err := clickup.NewClickUpConnector(clickUpAccounts)
	if err != nil {
		panic(err)
	}

	jiraAccounts, err := db.GetJiraAccounts(ctx.Ctx)
	if err != nil {
		panic(err)
	}
	jiraProvider, err := jira.NewJiraConnector(jiraAccounts)
	if err != nil {
		panic(err)
	}

	setCommands(&ctx, cfg, amqpProvider, clickupProvider, jiraProvider, logger, db)

	return &ctx, nil
}

func setCommands(
	ctx *Context,
	cfg *config.Config,
	queue *amqpwrapper.RabbitChannel,
	clickup *clickup.ConnectorPool,
	jira *jira.ConnectorPool,
	logger *logrus.Logger,
	db contract.Storage,
) {
	workerCmd := cmd.NewWorkerCmd(queue, clickup, jira)
	httpHandlerCmd := cmd.NewHTTPHandlerCmd(cfg, logger, queue, clickup, db)

	rootCmd := cmd.NewRootCmd()

	rootCmd.AddCommand(workerCmd)
	rootCmd.AddCommand(httpHandlerCmd)

	ctx.RootCmd = rootCmd
}

func (c *Context) Done() <-chan struct{} {
	return c.Ctx.Done()
}
