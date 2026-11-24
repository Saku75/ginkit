// Package v1 is version 1 of the API. Add a file per handler and register it
// here; paths are relative to /v1. The products, prices, and quotes handlers
// show common shapes: path and query parameters, a call through an injected
// client, and a validated body. Delete them when the real routes arrive.
package v1

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
)

// Rates is what the price handler needs from the rates client. A version
// declares the interfaces it consumes, so a fake can stand in for tests.
type Rates interface {
	Rate(ctx context.Context, from, to string) (float64, error)
}

func Register(api huma.API, fx Rates) {
	group := huma.NewGroup(api, "/v1")
	registerRoot(group)
	registerExample(group)
	registerProducts(group)
	registerPrices(group, fx)
	registerQuotes(group)
}
