package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	appcontext "x-qdo/jiraclick/context"
)

var (
	configPath = "."
)

func init() {
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		configPath = path
	}
}

func main() {
	ctx, err := appcontext.NewContext(configPath)
	if err != nil {
		panic(fmt.Errorf("context has thrown an error: %w", err))
	}
	defer ctx.CancelF()

	go waitShutdown(ctx.CancelF)

	if err := ctx.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	<-ctx.Done()
	ctx.WaitGroup.Wait()
}

func waitShutdown(cancelF context.CancelFunc) {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT)

	select {
	case s := <-sigint:
		fmt.Printf("os signal received: %[1]d (%[1]s)", s)
		cancelF()
	}
}
