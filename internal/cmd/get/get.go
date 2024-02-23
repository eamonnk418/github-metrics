package get

import (
	"github.com/eamonnk418/github-metrics/internal/config"
	"github.com/spf13/cobra"
)

func NewCmdGet(app *config.Application) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			app.Logger.InfoContext(cmd.Context(), "get command")
			return nil
		},
	}

	cmd.AddCommand(NewCmdGetRepo(app))

	return cmd
}