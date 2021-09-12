package main

import (
	"context"
	"flag"

	"kingim/examples/echo"
	"kingim/examples/kimbench"
	"kingim/examples/mock"
	"kingim/logger"
	"github.com/spf13/cobra"
)

const version = "v1"

func main() {
	flag.Parse()

	root := &cobra.Command{
		Use:     "kim",
		Version: version,
		Short:   "tools",
	}
	ctx := context.Background()

	// run echo client
	root.AddCommand(echo.NewCmd(ctx))

	// mock
	root.AddCommand(mock.NewClientCmd(ctx))
	root.AddCommand(mock.NewServerCmd(ctx))
	root.AddCommand(kimbench.NewBenchmarkCmd(ctx))

	if err := root.Execute(); err != nil {
		logger.WithError(err).Fatal("Could not run command")
	}
}
