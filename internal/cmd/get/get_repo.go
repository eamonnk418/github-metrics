package get

import (
	"fmt"

	"github.com/eamonnk418/github-metrics/internal/client"
	"github.com/eamonnk418/github-metrics/internal/config"
	"github.com/eamonnk418/github-metrics/internal/util"
	"github.com/spf13/cobra"
)

var opts = struct {
	owner string
	repo  string
}{}

func NewCmdGetRepo(app *config.Application) *cobra.Command {

	cmdGetRepo := &cobra.Command{
		Use:   "repo",
		Short: "Get a repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := util.ParseLocalCmdFlags(cmd, args); err != nil {
				return err
			}

			return getRepoAction(app, cmd, args)
		},
	}

	cmdGetRepo.Flags().StringVarP(&opts.owner, "owner", "o", "", "The owner of the repository")
	cmdGetRepo.Flags().StringVarP(&opts.repo, "repo", "r", "", "The name of the repository")

	return cmdGetRepo
}

func getRepoAction(app *config.Application, cmd *cobra.Command, args []string) error {
	gh := client.NewRepoClient()

	repo, err := gh.GetRepo(cmd.Context(), opts.owner, opts.repo)
	if err != nil {
		app.Logger.ErrorContext(cmd.Context(), "error getting repo", "error", err)
	}

	fmt.Printf("%v\n", repo)

	return nil
}
