package context

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"sync"
	"x-qdo/jiraclick/cmd"
	"x-qdo/jiraclick/pkg/config"
	"x-qdo/jiraclick/pkg/provider"
)

type Context struct {
	Ctx       context.Context
	CancelF   context.CancelFunc
	Config    config.Config
	RootCmd   *cobra.Command
	WaitGroup *sync.WaitGroup
}

func NewContext(configPath string) (*Context, error) {
	var (
		ctx       Context
		waitGroup *sync.WaitGroup
	)
	ctx.Ctx, ctx.CancelF = context.WithCancel(context.Background())

	cfg, err := config.NewConfig(configPath)
	if err != nil {
		panic(err)
	}

	logger := logrus.New()
	if cfg.Debug {
		logger.SetLevel(logrus.DebugLevel)
	}

	amqpProvider, err := provider.NewRabbitChannel(ctx.Ctx, waitGroup, cfg, logger)
	if err != nil {
		panic(err)
	}

	setCommands(&ctx, cfg, amqpProvider)

	return &ctx, nil
}

func setCommands(
	ctx *Context,
	cfg *config.Config,
	queue *provider.RabbitChannel,
) {
	workerCmd := cmd.NewWorkerCmd(cfg, queue)
	httpHandlerCmd := cmd.NewHttpHandlerCmd(cfg)

	rootCmd := cmd.NewRootCmd()

	rootCmd.AddCommand(workerCmd)
	workerCmd.AddCommand(httpHandlerCmd)

	ctx.RootCmd = rootCmd
}

func (c *Context) Done() <-chan struct{} {
	return c.Ctx.Done()
}
