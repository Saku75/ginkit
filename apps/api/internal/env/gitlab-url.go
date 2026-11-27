package env

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	gitLabURLKey     string = "GITLAB_URL"
	gitLabURLDefault string = "https://gitlab.com"
)

// GitLabURL is the GitLab instance behind the project routes, without /api/v4.
func GitLabURL() string {
	return lookup(gitLabURLKey, gitLabURLDefault, parseGitLabURL)
}

func parseGitLabURL(s string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(s))
	if err != nil {
		return "", err
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", fmt.Errorf("GitLab URL %q must be an absolute http(s) URL", s)
	}
	return strings.TrimRight(u.String(), "/"), nil
}
