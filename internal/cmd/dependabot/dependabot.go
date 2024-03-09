package dependabot

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
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
	Merged         int
	Closed         int
	Open           int
	MergedDuration time.Duration
	ClosedDuration time.Duration
	OpenDuration   time.Duration
}

func NewDependabotCmd(app *config.Application) *cobra.Command {
	opts := struct {
		org   string
		repo  string
		actor string
		state string
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

			dependabotStats, err := processRepositories(cmd.Context(), app.Config, repos, opts.actor, opts.state)
			if err != nil {
				return err
			}

			header := []string{"Repository", "Merged", "Closed", "Open", "ClosedDuration", "MergedDuration", "OpenDuration"}

			if err := writeToCSV(dependabotStats, header); err != nil {
				return err
			}

			return nil
		},
	}

	dependabotCmd.Flags().StringVarP(&opts.org, "org", "o", "actions", "The organization to list repositories for")
	dependabotCmd.Flags().StringVarP(&opts.repo, "repo", "r", "", "The repository to get metrics for")
	dependabotCmd.Flags().StringVarP(&opts.actor, "actor", "a", "dependabot[bot]", "The actor to filter pull requests by")
	dependabotCmd.Flags().StringVarP(&opts.state, "state", "s", "all", "The state to filter pull requests by")
	dependabotCmd.MarkFlagsRequiredTogether("org", "repo", "actor", "state")

	return dependabotCmd
}

func processRepositories(ctx context.Context, cfg *config.Config, repositories []*github.Repository, actor, state string) (*DependabotStats, error) {
	var repoList []*github.Repository
	var prList []*github.PullRequest
	var prStats []*PRStats

	for _, repo := range repositories {
		prs, err := listPRsForRepoByActorAndState(ctx, cfg, repo, actor, state)
		if err != nil {
			return nil, err
		}

		prList = append(prList, prs...)
		repoList = append(repoList, repo)

		stats, err := calculatePRStats(prs)
		if err != nil {
			return nil, err
		}

		prStats = append(prStats, stats)
	}

	dependabotStats := &DependabotStats{
		Repository:  repoList,
		PullRequest: prList,
		PRStats:     prStats,
	}

	return dependabotStats, nil
}

func calculatePRStats(prs []*github.PullRequest) (*PRStats, error) {
	var countMerged, countClosed, countOpen int
	var mergedDuration, closedDuration, openDuration time.Duration

	for _, pr := range prs {
		if pr.GetState() == "closed" {
			if !pr.GetMergedAt().IsZero() {
				mergedDuration += pr.GetMergedAt().Sub(pr.GetCreatedAt().Time)
				countMerged++
			}
			closedDuration += pr.GetClosedAt().Sub(pr.GetCreatedAt().Time)
			countClosed++
		}

		if pr.GetState() == "open" {
			openDuration += time.Since(pr.GetCreatedAt().Time)
			countOpen++
		}
	}

	var avgMerged, avgClosed, avgOpen time.Duration
	if countMerged > 0 {
		avgMerged = mergedDuration / time.Duration(countMerged)
	}
	if countClosed > 0 {
		avgClosed = closedDuration / time.Duration(countClosed)
	}
	if countOpen > 0 {
		avgOpen = openDuration / time.Duration(countOpen)
	}

	stats := &PRStats{
		Merged:         countMerged,
		Closed:         countClosed,
		Open:           countOpen,
		MergedDuration: avgMerged,
		ClosedDuration: avgClosed,
		OpenDuration:   avgOpen,
	}

	return stats, nil
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

func listPRsForRepoByActorAndState(ctx context.Context, cfg *config.Config, repository *github.Repository, actor, state string) ([]*github.PullRequest, error) {
	client := createNewGitHubClient(cfg.AccessToken)

	owner := repository.GetOwner().GetLogin()
	repo := repository.GetName()

	prs, resp, err := client.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{
		State: state,
		ListOptions: github.ListOptions{
			Page:    0,
			PerPage: 100,
		},
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return prs, nil
	}

	if resp.StatusCode == http.StatusNotModified {
		return nil, nil
	}

	if resp.StatusCode == http.StatusUnprocessableEntity {
		return nil, fmt.Errorf("unprocessable entity")
	}

	return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
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
			fmt.Sprintf("%s", prStat.ClosedDuration),
			fmt.Sprintf("%s", prStat.MergedDuration),
			fmt.Sprintf("%s", prStat.OpenDuration),
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
