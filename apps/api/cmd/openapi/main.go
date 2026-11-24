// Command openapi writes the OpenAPI document that the web app generates its
// client from. go generate at the module root keeps openapi.yaml current. The
// document has no server; clients set the mount point as their base URL.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"ginkit/api/internal/app"
	"ginkit/api/internal/handlers"
)

func main() {
	out := flag.String("o", "openapi.yaml", "file to write the document to")
	flag.Parse()

	if err := run(*out); err != nil {
		fmt.Fprintln(os.Stderr, "openapi:", err)
		os.Exit(1)
	}
}

func run(out string) error {
	_, api, err := app.New(context.Background(), app.Config{}, handlers.Dependencies{})
	if err != nil {
		return err
	}
	doc, err := api.OpenAPI().YAML()
	if err != nil {
		return err
	}
	return os.WriteFile(out, doc, 0o644)
}
