package client

import (
	"context"
	"net/http"
	"time"

	"github.com/google/go-github/v60/github"
)

type Client struct {
	*github.Client
}

func New(token string) *Client {
	httpClient := &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   10 * time.Second,
	}

	client := github.NewClient(httpClient).WithAuthToken(token)

	return &Client{Client: client}
}

func (c Client) GetRepositoryByOrg(ctx context.Context, org, repo string) (*github.Repository, error) {
	repository, _, err := c.Repositories.Get(ctx, org, repo)
	if err != nil {
		return nil, err
	}

	return repository, nil
}

func (c Client) ListRepositoriesByOrg(ctx context.Context, org string) ([]*github.Repository, error) {
	var allRepos []*github.Repository

	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	for {
		repos, resp, err := c.Repositories.ListByOrg(ctx, org, opts)
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
