package util

import (
	"net/url"
	"strings"
)

func ParseURL(rawUrl string) (string, string, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return "", "", err
	}

	org := strings.Split(parsedUrl.Path, "/")[1]
	repo := strings.Split(parsedUrl.Path, "/")[2]

	return org, repo, nil
}
