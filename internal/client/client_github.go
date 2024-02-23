package client

import (
	"context"
	"net/http"
	"os"

	"github.com/google/go-github/v59/github"
)

//go:generate mockgen -destination=../../mocks/github/mock_repo_reader.go -package=mock_github github.com/eamonnk418/github-metrics/internal/client RepoReader
type RepoReader interface {
	Get(ctx context.Context, owner string, name string) (*github.Repository, *github.Response, error)
}

type RepoClient struct {
	Service RepoReader
}

func NewRepoClient() *RepoClient {
	httpClient := &http.Client{
		Transport: http.DefaultTransport,
	}

	githubClient := github.NewClient(httpClient)
	service := githubClient.WithAuthToken(os.Getenv("GITHUB_TOKEN"))

	return &RepoClient{
		Service: service.Repositories,
	}
}

func (c RepoClient) GetRepo(ctx context.Context, owner string, name string) (*github.Repository, error) {
	repo, _, err := c.Service.Get(ctx, owner, name)
	if err != nil {
		return nil, err
	}
	return repo, err
}
