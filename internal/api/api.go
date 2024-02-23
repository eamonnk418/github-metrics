package api

import (
	"context"

	"github.com/google/go-github/v59/github"
)

type RepoService interface {
	GetRepository(ctx context.Context, owner, name string) (*github.Repository, *github.Response, error)
}