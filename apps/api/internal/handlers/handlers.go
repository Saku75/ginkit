// Package handlers holds the HTTP handlers. The probes live here, the docs
// page in docs, and routes in one subpackage per API version, starting with v1.
package handlers

import (
	"context"

	"github.com/danielgtaylor/huma/v2"

	"ginkit/api/internal/handlers/docs"
	"ginkit/api/internal/handlers/v1"
)

// Register mounts every handler: the probes and /version at the root, everything else under
// basePath. Cancelling lifetime makes /readyz answer 503 while the server
// drains. Pass one Check per dependency the service needs before taking traffic.
func Register(lifetime context.Context, api huma.API, basePath string, deps Dependencies, checks ...Check) {
	registerHealthz(api)
	registerReadyz(lifetime, api, checks)
	registerVersion(api)
	docs.Register(api, basePath)
	v1.Register(huma.NewGroup(api, basePath))
}
