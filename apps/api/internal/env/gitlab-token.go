package env

import "strings"

const (
	gitLabTokenKey     string = "GITLAB_TOKEN"
	gitLabTokenDefault string = ""
)

// GitLabToken authenticates requests to GitLab. Empty sends no credentials,
// which reaches public projects only.
func GitLabToken() string {
	return lookup(gitLabTokenKey, gitLabTokenDefault, func(s string) (string, error) {
		return strings.TrimSpace(s), nil
	})
}
