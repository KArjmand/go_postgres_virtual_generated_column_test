package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// Context returns a context that is canceled when the process receives
// an interrupt or termination signal
func Context() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		cancel()
	}()

	return ctx
}
