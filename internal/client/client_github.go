package client

import (
	"context"
	"net/http"
	"os"

	"github.com/google/go-github/v59/github"
)

type RepoReader interface {
	Get(ctx context.Context, owner string, name string) (*github.Repository, *github.Response, error)
}

type RepoClient struct {
	service RepoReader
}

func NewRepoClient() *RepoClient {
	httpClient := &http.Client{
		Transport: http.DefaultTransport,
	}

	githubClient := github.NewClient(httpClient)
	service := githubClient.WithAuthToken(os.Getenv("GITHUB_TOKEN"))

	return &RepoClient{
		service: service.Repositories,
	}
}

func (c RepoClient) GetRepo(ctx context.Context, owner string, name string) (*github.Repository, error) {
	repo, _, err := c.service.Get(ctx, owner, name)
	if err != nil {
		return nil, err
	}
	return repo, err
}
