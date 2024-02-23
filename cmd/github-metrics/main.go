package main

import (
	"log/slog"
	"os"

	"github.com/eamonnk418/github-metrics/internal/cmd"
	"github.com/eamonnk418/github-metrics/internal/config"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	app := &config.Application{
		Logger: logger,
	}

	cmd.Execute(app)
}
