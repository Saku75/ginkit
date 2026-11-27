// Package v1 is version 1 of the API. Add a file per handler and register it
// here; paths are relative to /v1.
package v1

import (
	"github.com/danielgtaylor/huma/v2"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// GitLab is what v1 uses of the GitLab client: the SDK services its handlers
// call. The SDK declares those as interfaces, so a test implements this with
// the mocks in gitlab.com/gitlab-org/api/client-go/testing. A version declares
// what it consumes, never the concrete client.
type GitLab interface {
	Projects() gitlab.ProjectsServiceInterface
}

func Register(api huma.API, gl GitLab) {
	group := huma.NewGroup(api, "/v1")
	registerRoot(group)
	registerExample(group)
	registerProjects(group, gl)
}
