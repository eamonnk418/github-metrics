package main

import (
	"log/slog"
	"os"

	"github.com/eamonnk418/github-metrics/internal/cmd"
	"github.com/eamonnk418/github-metrics/internal/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	app := &config.Application{
		Config: cfg,
		Logger: logger,
	}

	cmd.Execute(app)
}
