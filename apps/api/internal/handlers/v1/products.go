package v1

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"
)

// Product is the resource behind the example routes. The catalog is a fixed
// list so the examples need no storage.
type Product struct {
	ID    string   `json:"id" pattern:"^p-[0-9]+$" example:"p-1001"`
	Name  string   `json:"name" example:"Keyboard"`
	Price float64  `json:"price" minimum:"0" example:"89" doc:"Price in EUR."`
	Tags  []string `json:"tags,omitempty"`
}

var catalog = []Product{
	{ID: "p-1001", Name: "Keyboard", Price: 89, Tags: []string{"peripheral"}},
	{ID: "p-1002", Name: "Mouse", Price: 39, Tags: []string{"peripheral"}},
	{ID: "p-1003", Name: "Monitor", Price: 249, Tags: []string{"display"}},
	{ID: "p-1004", Name: "USB-C cable", Price: 12.5, Tags: []string{"cable", "peripheral"}},
}

func findProduct(id string) (Product, bool) {
	for _, p := range catalog {
		if p.ID == id {
			return p, true
		}
	}
	return Product{}, false
}

type listProductsInput struct {
	Tag    string `query:"tag" doc:"Only products carrying this tag."`
	Limit  int    `query:"limit" default:"20" minimum:"1" maximum:"100"`
	Offset int    `query:"offset" default:"0" minimum:"0"`
}

type listProductsOutput struct {
	Total int `header:"X-Total-Count" doc:"Matching products before paging."`
	Body  struct {
		Items []Product `json:"items"`
	}
}

type productIDInput struct {
	ID string `path:"id" pattern:"^p-[0-9]+$" example:"p-1001"`
}

type productOutput struct {
	Body Product
}

func registerProducts(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "v1-list-products",
		Method:      http.MethodGet,
		Path:        "/products",
		Summary:     "List products",
		Description: "Filters and pages through query parameters and reports the unpaged count in a header.",
		Tags:        []string{"v1"},
	}, func(_ context.Context, in *listProductsInput) (*listProductsOutput, error) {
		matches := make([]Product, 0, len(catalog))
		for _, p := range catalog {
			if in.Tag == "" || slices.Contains(p.Tags, in.Tag) {
				matches = append(matches, p)
			}
		}

		start := min(in.Offset, len(matches))
		end := min(start+in.Limit, len(matches))
		out := &listProductsOutput{Total: len(matches)}
		out.Body.Items = matches[start:end]
		return out, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "v1-get-product",
		Method:      http.MethodGet,
		Path:        "/products/{id}",
		Summary:     "Get a product",
		Tags:        []string{"v1"},
		Errors:      []int{http.StatusNotFound},
	}, func(_ context.Context, in *productIDInput) (*productOutput, error) {
		p, ok := findProduct(in.ID)
		if !ok {
			return nil, huma.Error404NotFound(fmt.Sprintf("product %s not found", in.ID))
		}
		return &productOutput{Body: p}, nil
	})
}
