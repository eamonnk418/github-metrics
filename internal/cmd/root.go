package cmd

import (
	"github.com/eamonnk418/github-metrics/internal/cmd/dependabot"
	"github.com/eamonnk418/github-metrics/internal/config"
	"github.com/spf13/cobra"
)

func NewRootCmd(app *config.Application) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "github-metrics",
		Short: "GitHub Metrics is a tool to gather metrics from GitHub",
		Long:  "GitHub Metrics is a tool to gather metrics from GitHub", // replace this with a heredoc or template
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	rootCmd.AddCommand(dependabot.NewDependabotCmd(app))

	return rootCmd
}

func Execute(app *config.Application) {
	cobra.CheckErr(NewRootCmd(app).Execute())
}
