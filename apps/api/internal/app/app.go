// Package app builds the router and the Huma API; the server and the OpenAPI
// export both start here.
package app

import (
	"context"
	"fmt"
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
	CORSAllowedOrigins []string
	BasePath           string
}

func ConfigFromEnv() Config {
	return Config{
		Debug:              env.Mode() == env.ModeDevelopment,
		TrustedProxies:     env.TrustedProxies(),
		CORSAllowedOrigins: env.CORSAllowedOrigins(),
		BasePath:           env.BasePath(),
	}
}

// New builds the router. Pass one handlers.Check per dependency the service
// needs before it can take traffic.
func New(lifetime context.Context, cfg Config, checks ...handlers.Check) (*gin.Engine, huma.API, error) {
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
	router.Use(sloggin.NewWithFilters(slog.Default(), sloggin.IgnorePath("/healthz", "/readyz")), gin.Recovery())

	if len(cfg.CORSAllowedOrigins) > 0 {
		c := cors.DefaultConfig()
		c.AllowOrigins = cfg.CORSAllowedOrigins
		c.AllowHeaders = append(c.AllowHeaders, "Authorization")
		router.Use(cors.New(c))
	}

	hc := huma.DefaultConfig("Ginkit API", "1.0.0")
	hc.DocsPath = "" // Served from embedded files instead; see handlers/docs.
	hc.OpenAPIPath = cfg.BasePath + "/openapi"
	hc.SchemasPath = cfg.BasePath + "/schemas"
	api := humagin.New(router, hc)
	handlers.Register(lifetime, api, cfg.BasePath, checks...)

	return router, api, nil
}
