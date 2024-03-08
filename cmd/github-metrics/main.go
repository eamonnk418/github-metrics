package main

import (
	"fmt"

	"github.com/eamonnk418/github-metrics/internal/util"
)

func main() {
	// cfg, err := config.LoadConfig()
	// if err != nil {
	// 	panic(err)
	// }

	// logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// app := &config.Application{
	// 	Config: cfg,
	// 	Logger: logger,
	// }

	// cmd.Execute(app)

	org, repo, err := util.ParseURL("https://github.com/eamonnk418/github-metrics")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Org: %s, Repo: %s\n", org, repo)
}
