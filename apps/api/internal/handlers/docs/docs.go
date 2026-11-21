// Package docs serves the API reference and the OpenAPI document behind it.
// The reference is built from files embedded in the binary, so it works
// without internet access; go generate fetches them into assets/ before the
// first build, see fetch.go for the source and version and generate.go at the
// module root for the order.
package docs

import (
	"embed"
	"net/http"
	"regexp"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

//go:embed assets/web-components.min.js assets/styles.min.css
var assets embed.FS

const page = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="referrer" content="no-referrer">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>{{title}}</title>
    <link rel="stylesheet" href="{{base}}/assets/styles.min.css">
    <script src="{{base}}/assets/web-components.min.js"></script>
  </head>
  <body style="height: 100vh;">
    <elements-api
      apiDescriptionUrl="{{base}}/openapi.yaml"
      router="hash"
      layout="sidebar"
      tryItCredentialsPolicy="same-origin"
    ></elements-api>
  </body>
</html>`

var (
	prefixPattern = regexp.MustCompile(`^(/[A-Za-z0-9._~-]+)+$`)
	hostPattern   = regexp.MustCompile(`^[A-Za-z0-9.-]+(:[0-9]+)?$`)
)

// Register serves the reference at basePath itself, or at / when it is empty,
// its assets under basePath/assets, and the OpenAPI document at
// basePath/openapi.yaml and .json. Links and the document's server URL honor
// X-Forwarded-Prefix, -Host and -Proto, which a reverse proxy sets when it
// strips a prefix such as /api, so "Try it" reaches the API through the proxy.
func Register(api huma.API, basePath string) {
	title := "API Reference"
	if info := api.OpenAPI().Info; info != nil && info.Title != "" {
		title = info.Title + " Reference"
	}

	csp := strings.Join([]string{
		"default-src 'none'",
		"base-uri 'none'",
		"connect-src 'self'",
		"form-action 'none'",
		"frame-ancestors 'none'",
		"sandbox allow-same-origin allow-scripts allow-popups allow-popups-to-escape-sandbox allow-downloads",
		"script-src 'self'",
		"style-src 'self' 'unsafe-inline'",
	}, "; ")

	pagePath := basePath
	if pagePath == "" {
		pagePath = "/"
	}

	adapter := api.Adapter()
	adapter.Handle(&huma.Operation{Method: http.MethodGet, Path: pagePath}, func(ctx huma.Context) {
		base := forwardedPrefix(ctx) + basePath
		html := strings.NewReplacer("{{title}}", title, "{{base}}", base).Replace(page)
		ctx.SetHeader("Content-Security-Policy", csp)
		ctx.SetHeader("Content-Type", "text/html; charset=utf-8")
		ctx.BodyWriter().Write([]byte(html))
	})
	serve(adapter, basePath+"/assets/styles.min.css", "styles.min.css", "text/css; charset=utf-8")
	serve(adapter, basePath+"/assets/web-components.min.js", "web-components.min.js", "text/javascript; charset=utf-8")

	adapter.Handle(&huma.Operation{Method: http.MethodGet, Path: basePath + "/openapi.yaml"}, func(ctx huma.Context) {
		doc := document(api, publicBase(ctx, basePath))
		body, err := doc.YAML()
		if err != nil {
			ctx.SetStatus(http.StatusInternalServerError)
			return
		}
		ctx.SetHeader("Content-Type", "application/vnd.oai.openapi+yaml")
		ctx.BodyWriter().Write(body)
	})
	adapter.Handle(&huma.Operation{Method: http.MethodGet, Path: basePath + "/openapi.json"}, func(ctx huma.Context) {
		doc := document(api, publicBase(ctx, basePath))
		body, err := doc.MarshalJSON()
		if err != nil {
			ctx.SetStatus(http.StatusInternalServerError)
			return
		}
		ctx.SetHeader("Content-Type", "application/vnd.oai.openapi+json")
		ctx.BodyWriter().Write(body)
	})
}

// document is a copy of the API's document with the server set to where the
// caller reached it, so the shared document that go generate writes stays
// free of any server.
func document(api huma.API, server string) *huma.OpenAPI {
	doc := *api.OpenAPI()
	doc.Servers = []*huma.Server{{URL: server}}
	return &doc
}

func publicBase(ctx huma.Context, basePath string) string {
	scheme := "http"
	if ctx.TLS() != nil {
		scheme = "https"
	}
	if p := ctx.Header("X-Forwarded-Proto"); p == "http" || p == "https" {
		scheme = p
	}
	host := ctx.Host()
	if h := ctx.Header("X-Forwarded-Host"); hostPattern.MatchString(h) {
		host = h
	}
	return scheme + "://" + host + forwardedPrefix(ctx) + basePath
}

func forwardedPrefix(ctx huma.Context) string {
	prefix := strings.TrimRight(ctx.Header("X-Forwarded-Prefix"), "/")
	if prefixPattern.MatchString(prefix) {
		return prefix
	}
	return ""
}

func serve(adapter huma.Adapter, path, name, contentType string) {
	data, err := assets.ReadFile("assets/" + name)
	if err != nil {
		panic(err)
	}
	adapter.Handle(&huma.Operation{Method: http.MethodGet, Path: path}, func(ctx huma.Context) {
		ctx.SetHeader("Content-Type", contentType)
		ctx.SetHeader("Cache-Control", "public, max-age=86400")
		ctx.BodyWriter().Write(data)
	})
}
