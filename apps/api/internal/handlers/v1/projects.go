package v1

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type Project struct {
	ID             int64      `json:"id"`
	Name           string     `json:"name"`
	Path           string     `json:"path" doc:"Namespace and project path, as in the URL."`
	WebURL         string     `json:"webUrl" format:"uri"`
	LastActivityAt *time.Time `json:"lastActivityAt,omitempty"`
}

type listProjectsInput struct {
	Search string `query:"search" maxLength:"100" doc:"Matches project names and paths."`
	Limit  int64  `query:"limit" default:"20" minimum:"1" maximum:"100"`
}

type listProjectsOutput struct {
	Body struct {
		Items []Project `json:"items"`
	}
}

type projectIDInput struct {
	ID int64 `path:"id" minimum:"1"`
}

type projectOutput struct {
	Body Project
}

func registerProjects(api huma.API, gl GitLab) {
	huma.Register(api, huma.Operation{
		OperationID: "v1-list-projects",
		Method:      http.MethodGet,
		Path:        "/projects",
		Summary:     "List GitLab projects",
		Description: "Projects the configured token can see, most recently active first.",
		Tags:        []string{"v1"},
		Errors:      []int{http.StatusBadGateway},
	}, func(ctx context.Context, in *listProjectsInput) (*listProjectsOutput, error) {
		opts := &gitlab.ListProjectsOptions{
			ListOptions: gitlab.ListOptions{PerPage: in.Limit},
			OrderBy:     gitlab.Ptr("last_activity_at"),
			Simple:      gitlab.Ptr(true),
		}
		if in.Search != "" {
			opts.Search = gitlab.Ptr(in.Search)
		}
		found, _, err := gl.Projects().ListProjects(opts, gitlab.WithContext(ctx))
		if err != nil {
			return nil, gitlabError(ctx, err, "")
		}
		out := &listProjectsOutput{}
		out.Body.Items = make([]Project, 0, len(found))
		for _, p := range found {
			out.Body.Items = append(out.Body.Items, project(p))
		}
		return out, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "v1-get-project",
		Method:      http.MethodGet,
		Path:        "/projects/{id}",
		Summary:     "One GitLab project",
		Tags:        []string{"v1"},
		Errors:      []int{http.StatusNotFound, http.StatusBadGateway},
	}, func(ctx context.Context, in *projectIDInput) (*projectOutput, error) {
		p, _, err := gl.Projects().GetProject(in.ID, nil, gitlab.WithContext(ctx))
		if err != nil {
			return nil, gitlabError(ctx, err, "project not found")
		}
		return &projectOutput{Body: project(p)}, nil
	})
}

func project(p *gitlab.Project) Project {
	return Project{
		ID:             p.ID,
		Name:           p.Name,
		Path:           p.PathWithNamespace,
		WebURL:         p.WebURL,
		LastActivityAt: p.LastActivityAt,
	}
}

// gitlabError maps the SDK's error to the API's. GitLab's 404 is passed on as
// notFound when set; everything else is GitLab's problem, so the client gets
// a 502 and the log gets the detail. The SDK returns 404 as gitlab.ErrNotFound
// and every other status as a *gitlab.ErrorResponse.
func gitlabError(ctx context.Context, err error, notFound string) error {
	if notFound != "" && errors.Is(err, gitlab.ErrNotFound) {
		return huma.Error404NotFound(notFound)
	}
	var resp *gitlab.ErrorResponse
	if errors.As(err, &resp) && resp.Response != nil {
		switch resp.Response.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			slog.ErrorContext(ctx, "gitlab rejected the token", slog.Any("error", err))
			return huma.Error502BadGateway("GitLab rejected the configured token")
		}
	}
	slog.ErrorContext(ctx, "gitlab request failed", slog.Any("error", err))
	return huma.Error502BadGateway("GitLab did not answer")
}
