// Package docs serves the API reference from files embedded in the binary, so
// it works without internet access. go generate fetches them into assets/
// before the first build; see fetch.go for the source and version.
package docs

import (
	"embed"
	"net/http"
	"regexp"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

//go:generate go run fetch.go

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

var prefixPattern = regexp.MustCompile(`^(/[A-Za-z0-9._~-]+)+$`)

// Register serves the docs page at basePath itself, or at / when it is empty,
// with its assets under basePath/assets. Links honor X-Forwarded-Prefix, which
// a reverse proxy sets when it strips a prefix such as /api before forwarding.
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
