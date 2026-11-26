// Package app builds the router and the Huma API; the server and the OpenAPI
// export both start here.
package app

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"io"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"

	"ginkit/api/internal/env"
	"ginkit/api/internal/handlers"
)

// Config selects what New builds. The zero value is release mode with nothing
// mounted under a prefix.
type Config struct {
	Debug              bool
	TrustedProxies     []string
	ClientIPHeader     string
	CORSAllowedOrigins []string
	BasePath           string
}

// jsonFormat reads and writes bodies with encoding/json/v2. Besides stricter
// decoding, it lets a struct keep properties it does not declare in a map
// tagged `json:",embed"`; v1.Customer shows the pattern.
var jsonFormat = huma.Format{
	Marshal:   func(w io.Writer, v any) error { return json.MarshalWrite(w, v) },
	Unmarshal: func(data []byte, v any) error { return json.Unmarshal(data, v) },
}

func ConfigFromEnv() Config {
	return Config{
		Debug:              env.Mode() == env.ModeDevelopment,
		TrustedProxies:     env.TrustedProxies(),
		ClientIPHeader:     env.ClientIPHeader(),
		CORSAllowedOrigins: env.CORSAllowedOrigins(),
		BasePath:           env.BasePath(),
	}
}

// New builds the router. deps carries the clients handlers call; pass one
// handlers.Check per dependency the service needs before it can take traffic.
func New(lifetime context.Context, cfg Config, deps handlers.Dependencies, checks ...handlers.Check) (*gin.Engine, huma.API, error) {
	// gin.SetMode must run before gin.New.
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	if err := router.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		return nil, nil, fmt.Errorf("set trusted proxies: %w", err)
	}
	router.TrustedPlatform = cfg.ClientIPHeader
	// Probe and version requests are noise unless the probes themselves are
	// being debugged, so they are logged only at LOG_LEVEL=debug.
	var filters []sloggin.Filter
	if !slog.Default().Enabled(lifetime, slog.LevelDebug) {
		filters = append(filters, sloggin.IgnorePath("/healthz", "/readyz", "/version"))
	}
	router.Use(sloggin.NewWithFilters(slog.Default(), filters...), gin.Recovery())

	if len(cfg.CORSAllowedOrigins) > 0 {
		c := cors.DefaultConfig()
		c.AllowOrigins = cfg.CORSAllowedOrigins
		c.AllowHeaders = append(c.AllowHeaders, "Authorization")
		router.Use(cors.New(c))
	}

	hc := huma.DefaultConfig("Ginkit API", "1.0.0")
	hc.Formats = map[string]huma.Format{"application/json": jsonFormat, "json": jsonFormat}
	hc.DocsPath = ""    // Both are served by handlers/docs, which knows the
	hc.OpenAPIPath = "" // public base URL behind a proxy.
	// Declare $schema on every response schema, so JSON files that point at
	// /schemas validate in editors, but leave it out of responses: the link
	// could not know the public URL behind a proxy.
	hc.CreateHooks = []func(huma.Config) huma.Config{func(c huma.Config) huma.Config {
		links := huma.NewSchemaLinkTransformer("#/components/schemas/", c.SchemasPath)
		c.OnAddOperation = append(c.OnAddOperation, links.OnAddOperation)
		return c
	}}
	hc.SchemasPath = cfg.BasePath + "/schemas"
	api := humagin.New(router, hc)
	handlers.Register(lifetime, api, cfg.BasePath, deps, checks...)

	return router, api, nil
}
