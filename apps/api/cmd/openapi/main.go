// Command openapi writes the OpenAPI document that the web app generates its
// client from. go generate keeps openapi.yaml current. The document has no
// base path; clients set the mount point as their base URL.
package main

//go:generate go run . -o ../../openapi.yaml

import (
	"context"
	"flag"
	"fmt"
	"os"

	"ginkit/api/internal/app"
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
	_, api, err := app.New(context.Background(), app.Config{})
	if err != nil {
		return err
	}
	doc, err := api.OpenAPI().YAML()
	if err != nil {
		return err
	}
	return os.WriteFile(out, doc, 0o644)
}
