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

type DependabotStats struct {
	Repository  []*github.Repository
	PullRequest []*github.PullRequest
	PRStats     []*PRStats
}

type PRStats struct {
	Merged      int
	Closed      int
	Open        int
	AvgDuration time.Duration
}

func NewDependabotCmd(app *config.Application) *cobra.Command {
	opts := struct {
		org  string
		repo string
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

			var repoList []*github.Repository
			var prList []*github.PullRequest

			for _, r := range repos {
				repo, err := getRepository(cmd.Context(), app.Config, opts.org, r.GetName())
				if err != nil {
					return err
				}

				repoList = append(repoList, repo)
			}

			var prStats []*PRStats
			for _, repo := range repoList {
				prs, err := listPrsForRepo(cmd.Context(), app.Config, opts.org, repo.GetName(), "dependabot[bot]")
				if err != nil {
					return err
				}
				prList = append(prList, prs...)

				var durations []time.Duration
				closedCount := 0
				openCount := 0
				mergedCount := 0

				for _, pr := range prs {
					if pr.GetUser().GetLogin() != "dependabot[bot]" {
						continue
					}

					switch pr.GetState() {
					case "closed":
						closedCount++
						if pr.GetMerged() {
							// Calculate and store the duration for merged PRs
							mergedAt := pr.GetMergedAt()
							createdAt := pr.GetCreatedAt()

							duration := mergedAt.Sub(createdAt.Time)
							durations = append(durations, duration)
						}
					case "open":
						openCount++
						createdAt := pr.GetCreatedAt()

						duration := time.Since(createdAt.Time)
						durations = append(durations, duration)
					}

					if pr.GetMerged() {
						mergedCount++
						mergedAt := pr.GetMergedAt()
						createdAt := pr.GetCreatedAt()

						duration := mergedAt.Sub(createdAt.Time)
						durations = append(durations, duration)
					}
				}

				// Calculate the average duration for the repository
				var totalDuration time.Duration
				for _, duration := range durations {
					totalDuration += duration
				}

				var avgDuration time.Duration
				if len(durations) > 0 {
					avgDuration = totalDuration / time.Duration(len(durations))
				}

				prStats = append(prStats, &PRStats{
					Merged:      mergedCount,
					Closed:      closedCount,
					Open:        openCount,
					AvgDuration: avgDuration,
				})

				fmt.Printf("Repository: %s, Merged: %d, Closed: %d, Opened: %d, Duration(Avg): %s\n", repo.GetFullName(), mergedCount, closedCount, openCount, avgDuration)
			}

			data := &DependabotStats{
				Repository:  repoList,
				PullRequest: prList,
				PRStats:     prStats,
			}

			header := []string{"Repository", "Merged", "Closed", "Open", "AvgDuration"}
			if err := writeToCSV(data, header); err != nil {
				return err
			}

			return nil
		},
	}

	dependabotCmd.Flags().StringVarP(&opts.org, "org", "o", "", "The organization to list repositories for")
	dependabotCmd.Flags().StringVarP(&opts.repo, "repo", "r", "", "The repository to get metrics for")
	dependabotCmd.MarkFlagsOneRequired("org", "repo")

	return dependabotCmd
}

func getRepository(ctx context.Context, cfg *config.Config, org, repo string) (*github.Repository, error) {
	gh := createNewGitHubClient(cfg.AccessToken)

	repository, err := gh.GetRepositoryByOrg(ctx, org, repo)
	if err != nil {
		return nil, err
	}

	return repository, nil
}

func listRepositories(ctx context.Context, cfg *config.Config, org string) ([]*github.Repository, error) {
	gh := createNewGitHubClient(cfg.AccessToken)

	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	var allRepos []*github.Repository

	for {
		repos, resp, err := gh.Repositories.ListByOrg(ctx, org, opts)
		if err != nil {
			return nil, err
		}

		allRepos = append(allRepos, repos...)

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

func listPrsForRepo(ctx context.Context, cfg *config.Config, org, repo, actor string) ([]*github.PullRequest, error) {
	gh := createNewGitHubClient(cfg.AccessToken)

	opts := &github.PullRequestListOptions{
		State: "all",
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	var allPrs []*github.PullRequest

	for {
		prs, resp, err := gh.PullRequests.List(ctx, org, repo, opts)
		if err != nil {
			return nil, err
		}

		allPrs = append(allPrs, prs...)

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	return allPrs, nil
}

func writeToCSV(data *DependabotStats, header []string) error {
	csvFile, err := os.Create(fmt.Sprintf("%s-%s.csv", "repos", time.Now().Format("2006-01-02")))
	if err != nil {
		return err
	}
	defer csvFile.Close()

	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	// Write header row
	if err := csvWriter.Write(header); err != nil {
		return err
	}

	// Iterate over repositories and their associated PRStats
	for i, repo := range data.Repository {
		prStat := data.PRStats[i]

		// Write a row for each repository and its associated PRStats
		if err := csvWriter.Write([]string{
			repo.GetFullName(),
			fmt.Sprintf("%d", prStat.Merged),
			fmt.Sprintf("%d", prStat.Closed),
			fmt.Sprintf("%d", prStat.Open),
			fmt.Sprintf("%s", prStat.AvgDuration),
		}); err != nil {
			return err
		}
	}

	return nil
}

func createNewGitHubClient(token string) *client.Client {
	return client.New(token)
}








// // ... (import statements)

// // DependabotStats and PRStats structures remain the same

// func NewDependabotCmd(app *config.Application) *cobra.Command {
// 	opts := struct {
// 		org  string
// 		repo string
// 	}{}

// 	dependabotCmd := &cobra.Command{
// 		Use:   "dependabot",
// 		Short: "Dependabot is a tool to gather metrics from Dependabot",
// 		Long:  "Dependabot is a tool to gather metrics from Dependabot. It lists repositories and writes metrics to a CSV file.",
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			repos, err := listRepositories(cmd.Context(), app.Config, opts.org)
// 			if err != nil {
// 				return err
// 			}

// 			repoList, prList, prStats, err := processRepositories(cmd.Context(), app.Config, opts.org, repos)
// 			if err != nil {
// 				return err
// 			}

// 			data := &DependabotStats{
// 				Repository:  repoList,
// 				PullRequest: prList,
// 				PRStats:     prStats,
// 			}

// 			header := []string{"Repository", "Merged", "Closed", "Open", "AvgDuration"}
// 			if err := writeToCSV(data, header); err != nil {
// 				return err
// 			}

// 			return nil
// 		},
// 	}

// 	// ... (flag definitions)

// 	return dependabotCmd
// }

// func processRepositories(ctx context.Context, cfg *config.Config, org string, repos []*github.Repository) ([]*github.Repository, []*github.PullRequest, []*PRStats, error) {
// 	var repoList []*github.Repository
// 	var prList []*github.PullRequest
// 	var prStats []*PRStats

// 	for _, repo := range repos {
// 		prs, err := listPrsForRepo(ctx, cfg, org, repo.GetName(), "dependabot[bot]")
// 		if err != nil {
// 			return nil, nil, nil, err
// 		}
// 		prList = append(prList, prs...)

// 		repoList = append(repoList, repo)

// 		stats, err := calculatePRStats(prs)
// 		if err != nil {
// 			return nil, nil, nil, err
// 		}
// 		prStats = append(prStats, stats)

// 		printRepoStats(repo, stats)
// 	}

// 	return repoList, prList, prStats, nil
// }

// func calculatePRStats(prs []*github.PullRequest) (*PRStats, error) {
// 	var durations []time.Duration
// 	closedCount := 0
// 	openCount := 0
// 	mergedCount := 0

// 	for _, pr := range prs {
// 		if pr.GetUser().GetLogin() != "dependabot[bot]" {
// 			continue
// 		}

// 		switch pr.GetState() {
// 		case "closed":
// 			closedCount++
// 			if pr.GetMerged() {
// 				duration, err := calculateDuration(pr.GetMergedAt(), pr.GetCreatedAt())
// 				if err != nil {
// 					return nil, err
// 				}
// 				durations = append(durations, duration)
// 			}
// 		case "open":
// 			openCount++
// 			duration, err := calculateDuration(time.Now(), pr.GetCreatedAt())
// 			if err != nil {
// 				return nil, err
// 			}
// 			durations = append(durations, duration)
// 		}

// 		if pr.GetMerged() {
// 			mergedCount++
// 			duration, err := calculateDuration(pr.GetMergedAt(), pr.GetCreatedAt())
// 			if err != nil {
// 				return nil, err
// 			}
// 			durations = append(durations, duration)
// 		}
// 	}

// 	// Calculate the average duration for the repository
// 	avgDuration, err := calculateAvgDuration(durations)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &PRStats{
// 		Merged:      mergedCount,
// 		Closed:      closedCount,
// 		Open:        openCount,
// 		AvgDuration: avgDuration,
// 	}, nil
// }

// func calculateDuration(end, start *github.Timestamp) (time.Duration, error) {
// 	if end == nil || start == nil {
// 		return 0, errors.New("timestamp is nil")
// 	}

// 	return end.Time.Sub(start.Time), nil
// }

// func calculateAvgDuration(durations []time.Duration) (time.Duration, error) {
// 	var totalDuration time.Duration
// 	for _, duration := range durations {
// 		totalDuration += duration
// 	}

// 	if len(durations) > 0 {
// 		return totalDuration / time.Duration(len(durations)), nil
// 	}

// 	return 0, errors.New("cannot calculate average duration with no durations")
// }

// func printRepoStats(repo *github.Repository, stats *PRStats) {
// 	fmt.Printf("Repository: %s, Merged: %d, Closed: %d, Opened: %d, Duration(Avg): %s\n",
// 		repo.GetFullName(), stats.Merged, stats.Closed, stats.Open, stats.AvgDuration)
// }

// // ... (writeToCSV function remains the same)
