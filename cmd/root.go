package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use: "jiraclick",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Please use `--help` for more details")
		},
	}
}
