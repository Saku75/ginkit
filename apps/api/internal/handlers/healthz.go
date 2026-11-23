package handlers

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"ginkit/api/internal/version"
)

type healthzOutput struct {
	Body struct {
		Status  string `json:"status" enum:"ok" doc:"Always ok while the process is serving requests."`
		Version string `json:"version" doc:"Build version, the same for the API and web images of one release."`
	}
}

func registerHealthz(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "healthz",
		Method:      http.MethodGet,
		Path:        "/healthz",
		Summary:     "Liveness probe",
		Description: "Returns 200 whenever the process is up and able to serve requests.",
		Tags:        []string{"Health"},
	}, func(_ context.Context, _ *struct{}) (*healthzOutput, error) {
		out := &healthzOutput{}
		out.Body.Status = "ok"
		out.Body.Version = version.Version
		return out, nil
	})
}
