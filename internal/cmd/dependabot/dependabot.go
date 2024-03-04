package dependabot

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/eamonnk418/github-metrics/internal/client"
	"github.com/eamonnk418/github-metrics/internal/config"
	"github.com/google/go-github/v60/github"
	"github.com/spf13/cobra"
)

func NewDependabotCmd(app *config.Application) *cobra.Command {
	opts := struct {
		org string
	}{}

	dependabotCmd := &cobra.Command{
		Use:   "dependabot",
		Short: "Dependabot is a tool to gather metrics from Dependabot",
		Long:  "Dependabot is a tool to gather metrics from Dependabot. It lists repositories and writes metrics to a CSV file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			repos, err := listRepositories(cmd.Context(), app.Config, opts.org)
			if err != nil {
				return err
			}

			if err := writeToCSV(repos); err != nil {
				return err
			}

			return nil
		},
	}

	dependabotCmd.Flags().StringVarP(&opts.org, "org", "o", "", "The organization to list repositories for")
	if err := dependabotCmd.MarkFlagRequired("org"); err != nil {
		cobra.CheckErr(err)
	}

	return dependabotCmd
}

func listRepositories(ctx context.Context, cfg *config.Config, org string) ([]*github.Repository, error) {
	gh := client.New(cfg.AccessToken)

	repos, err := gh.ListRepositoriesByOrg(ctx, org)
	if err != nil {
		return nil, err
	}

	return repos, nil
}

func writeToCSV(repos []*github.Repository) error {
	csvFile, err := os.Create(fmt.Sprintf("%s-%s.csv", "repos", time.Now().Format("2006-01-02")))
	if err != nil {
		return err
	}
	defer csvFile.Close()

	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	if err := csvWriter.Write([]string{"Repository", "Stars", "Watchers", "Forks", "OpenIssues", "Size"}); err != nil {
		return err
	}

	for _, repo := range repos {
		if err := csvWriter.Write([]string{
			repo.GetFullName(),
			fmt.Sprint(repo.GetStargazersCount()),
			fmt.Sprint(repo.GetWatchersCount()),
			fmt.Sprint(repo.GetForksCount()),
			fmt.Sprint(repo.GetOpenIssuesCount()),
			fmt.Sprint(repo.GetSize()),
		}); err != nil {
			return err
		}
	}

	return nil
}
