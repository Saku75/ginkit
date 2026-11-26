package handlers

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"ginkit/api/internal/version"
)

type versionOutput struct {
	Body struct {
		Version string `json:"version" doc:"Build version, the same for the API and web images of one release."`
	}
}

func registerVersion(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "version",
		Method:      http.MethodGet,
		Path:        "/version",
		Summary:     "Build version",
		Description: "The version this binary was built with. The web app's /readyz compares it with its own.",
		Tags:        []string{"Health"},
	}, func(_ context.Context, _ *struct{}) (*versionOutput, error) {
		out := &versionOutput{}
		out.Body.Version = version.Version
		return out, nil
	})
}
