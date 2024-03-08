package util_test

import (
	"testing"

	"github.com/eamonnk418/github-metrics/internal/util"
)

func TestUrlParser(t *testing.T) {
	tests := []struct {
		name string
		url  string
		org  string
		repo string
		err  error
	}{
		{
			name: "ValidURL",
			url:  "https://github.com/test-org/test-repo",
			org:  "test-org",
			repo: "test-repo",
			err:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			org, repo, err := util.ParseURL(tt.url)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if org != tt.org {
				t.Errorf("expected org: %s, got: %s", tt.org, org)
			}

			if repo != tt.repo {
				t.Errorf("expected repo: %s, got: %s", tt.repo, repo)
			}
		})
	}
}
