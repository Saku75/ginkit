// Package gitlab builds the client for gitlab.com/gitlab-org/api/client-go once
// for cmd/api/main.go and adds the readiness check the SDK lacks. Handlers do
// not take this type: each API version declares an interface over the SDK
// services it calls, which Client satisfies with one accessor per service.
package gitlab

import (
	"context"
	"fmt"
	"net/http"

	gl "gitlab.com/gitlab-org/api/client-go"
)

type Client struct {
	api *gl.Client
}

// New points the SDK at one GitLab instance, given without /api/v4. An empty
// token sends no credentials, which reaches public projects only. Timeouts
// belong to httpClient; the SDK retries failed requests on its own.
func New(baseURL, token string, httpClient *http.Client) (*Client, error) {
	api, err := gl.NewClient(token, gl.WithBaseURL(baseURL), gl.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("gitlab: %w", err)
	}
	return &Client{api: api}, nil
}

// Projects is the SDK's projects service. Add an accessor per service handlers use.
func (c *Client) Projects() gl.ProjectsServiceInterface { return c.api.Projects }

// Ping lists one project, which works with and without a token where GET
// /version does not. It fails when GitLab is unreachable or rejects the token.
// The SDK's retries are off for this one request: they take over ten seconds
// to give up on an outage, and a probe is repeated anyway.
func (c *Client) Ping(ctx context.Context) error {
	opts := &gl.ListProjectsOptions{ListOptions: gl.ListOptions{PerPage: 1}, Simple: gl.Ptr(true)}
	_, _, err := c.api.Projects.ListProjects(opts, gl.WithContext(ctx), gl.WithRequestRetry(noRetry))
	if err != nil {
		return fmt.Errorf("gitlab: %w", err)
	}
	return nil
}

func noRetry(context.Context, *http.Response, error) (bool, error) { return false, nil }
