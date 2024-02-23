package client_test

import (
	"context"
	"testing"

	"github.com/eamonnk418/github-metrics/internal/client"
	mock_github "github.com/eamonnk418/github-metrics/mocks/github"
	"github.com/google/go-github/v59/github"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetRepo(t *testing.T) {
	tests := []struct {
		name     string
		owner    string
		repo     string
		wantRepo *github.Repository
		err      error
	}{
		{
			name:     "success",
			owner:    "example-test-owner",
			repo:     "example-test-repo",
			wantRepo: &github.Repository{},
			err:      nil,
		},
		{
			name:     "error",
			owner:    "example-test-owner",
			repo:     "example-test-repo",
			wantRepo: nil,
			err:      assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockedRepoReader := mock_github.NewMockRepoReader(ctrl)

			mockedRepoReader.EXPECT().Get(gomock.Any(), tt.owner, tt.repo).Return(tt.wantRepo, nil, tt.err).Times(1)

			repoClient := &client.RepoClient{Service: mockedRepoReader}

			repo, err := repoClient.GetRepo(context.Background(), tt.owner, tt.repo)
			if tt.err != nil {
				assert.Error(t, err)
			}

			assert.Equal(t, tt.wantRepo, repo)
		})
	}
}
