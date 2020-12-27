package cmd

import (
	"github.com/spf13/cobra"

	"x-qdo/jiraclick/pkg/consumer"
	"x-qdo/jiraclick/pkg/contract"
	"x-qdo/jiraclick/pkg/provider"
	"x-qdo/jiraclick/pkg/provider/clickup"
	"x-qdo/jiraclick/pkg/provider/jira"
)

func NewWorkerCmd(
	queue *provider.RabbitChannel,
	clickup *clickup.APIClient,
	jira *jira.ConnectorPool,
) *cobra.Command {
	return &cobra.Command{
		Use:   "worker",
		Short: "Runs tasks consumer",
		Long:  `Runs consumer to receive and process tasks from queue.`,
		Run: func(cmd *cobra.Command, args []string) {
			var (
				cons contract.Consumer
				err  error
			)

			go func() {
				cons, err = consumer.NewActionsConsumer(jira, queue, clickup)
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
