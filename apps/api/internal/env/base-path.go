package env

import (
	"fmt"
	"strings"
)

const (
	basePathKey     string = "BASE_PATH"
	basePathDefault string = ""
)

// BasePath is the URL prefix everything but the probes is mounted under.
func BasePath() string {
	return lookup(basePathKey, basePathDefault, parseBasePath)
}

func parseBasePath(s string) (string, error) {
	s = strings.TrimRight(strings.TrimSpace(s), "/")
	if s == "" {
		return "", nil
	}
	if !strings.HasPrefix(s, "/") {
		return "", fmt.Errorf("base path %q must start with /", s)
	}
	return s, nil
}
