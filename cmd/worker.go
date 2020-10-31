package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"x-qdo/jiraclick/pkg/config"
	"x-qdo/jiraclick/pkg/consumer"
	"x-qdo/jiraclick/pkg/contract"
	"x-qdo/jiraclick/pkg/provider"
)

func NewWorkerCmd(
	cfg *config.Config,
	queue *provider.RabbitChannel,
) *cobra.Command {
	return &cobra.Command{
		Use:   "worker [OBJECT]",
		Short: "Runs tasks consumer",
		Long:  `Runs consumer to receive and process tasks from queue.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires an object argument")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			var (
				cons contract.Consumer
				err  error
			)

			go func() {
				cons, err = consumer.NewActionsConsumer(cfg, queue)
				if err != nil {
					panic(err)
				}

				err = cons.SetUpListeners()
				if err != nil {
					panic(err)
				}
			}()
		},
	}
}
