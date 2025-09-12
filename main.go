package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cnoe-io/idpbuilder/pkg/cmd"
)

func main() {
	interrupted := make(chan os.Signal, 1)
	defer close(interrupted)
	signal.Notify(interrupted, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupted)

	ctx, cancel := context.WithCancelCause(context.Background())

	go func() {
		select {
		case <-interrupted:
			cancel(fmt.Errorf("command interrupted"))
		}
	}()

	cmd.Execute(ctx)
}
