package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	appcontext "x-qdo/jiraclick/context"
)

func main() {
	ctx, err := appcontext.NewContext()
	if err != nil {
		panic(fmt.Errorf("context has thrown an error: %w", err))
	}
	defer ctx.CancelF()

	go waitShutdown(ctx.CancelF)

	if err := ctx.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		ctx.CancelF()
	}
	<-ctx.Done()
	ctx.WaitGroup.Wait()
}

func waitShutdown(cancelF context.CancelFunc) {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT)

	s := <-sigint
	fmt.Printf("os signal received: %[1]d (%[1]s)", s)
	cancelF()
}
