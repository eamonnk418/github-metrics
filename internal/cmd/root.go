package cmd

import (
	"github.com/eamonnk418/github-metrics/internal/cmd/get"
	"github.com/eamonnk418/github-metrics/internal/config"
	"github.com/spf13/cobra"
)

func NewRootCmd(app *config.Application) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "github-metrics",
		Short: "A CLI for interacting with GitHub",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	rootCmd.AddCommand(get.NewCmdGet(app))

	return rootCmd
}

func Execute(app *config.Application) {
	cobra.CheckErr(NewRootCmd(app).Execute())
}
