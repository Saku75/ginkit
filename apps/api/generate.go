// Package api holds the module's go generate directives. They run in this
// order on purpose: the openapi command embeds the docs assets, so those have
// to be fetched first.
package api

//go:generate go run ./internal/handlers/docs/fetch.go -o internal/handlers/docs/assets
//go:generate go run ./cmd/openapi -o openapi.yaml
