package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/belo4ya/runy"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run() error {
	ctx := runy.SetupSignalHandler(context.Background())
	_ = ctx
	return nil
}
