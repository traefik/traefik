package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// ContextWithSignal create a context cancelled when SIGINT or SIGTERM are notified
func ContextWithSignal(ctx context.Context) context.Context {
	newCtx, cancel := context.WithCancel(ctx)
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-signals:
			cancel()
		}
	}()
	return newCtx
}
