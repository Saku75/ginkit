package v1

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

// QuoteRequest shows the three ways a body can treat properties it does not
// declare. The request itself is strict, which is Huma's default for structs:
// an unknown property fails validation. Lines takes only the QuoteLine shape.
// Customer accepts extra properties and keeps them; see its Extra field. Notes
// takes any object at all.
type QuoteRequest struct {
	Lines    []QuoteLine    `json:"lines" minItems:"1" maxItems:"50"`
	Customer Customer       `json:"customer"`
	Shipping string         `json:"shipping,omitempty" enum:"standard,express" default:"standard"`
	Coupon   string         `json:"coupon,omitempty" pattern:"^[A-Z0-9]{4,12}$" doc:"SAVE10 is the only code the example knows."`
	Notes    map[string]any `json:"notes,omitempty" doc:"Any object the client wants to attach."`
}

type QuoteLine struct {
	ProductID string `json:"productId" pattern:"^p-[0-9]+$" example:"p-1001"`
	Quantity  int    `json:"quantity" minimum:"1" maximum:"999" example:"2"`
}

// Customer allows properties it does not declare, through the `_` tag, and
// keeps them: encoding/json/v2 fills the embedded Extra map with every unknown
// member and writes them back on output, so the object survives a round trip
// with modifications. hidden keeps Extra out of the schema. Without such a
// field, unknown properties pass validation and are dropped on decoding.
type Customer struct {
	_       struct{}       `json:"-" additionalProperties:"true"`
	Name    string         `json:"name" minLength:"1" maxLength:"100" example:"Ada Lovelace"`
	Email   string         `json:"email" format:"email" example:"ada@example.com"`
	Country string         `json:"country" pattern:"^[A-Z]{2}$" example:"DK" doc:"ISO 3166-1 alpha-2 code."`
	Extra   map[string]any `json:",embed" hidden:"true"`
}

type Quote struct {
	Currency  string      `json:"currency" enum:"EUR"`
	Customer  Customer    `json:"customer" doc:"The request's customer, with the email normalized and any extra properties kept."`
	Lines     []QuoteItem `json:"lines"`
	Subtotal  float64     `json:"subtotal"`
	Discount  float64     `json:"discount"`
	Shipping  float64     `json:"shipping"`
	Total     float64     `json:"total"`
	ExpiresAt time.Time   `json:"expiresAt"`
}

type QuoteItem struct {
	ProductID string  `json:"productId"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unitPrice"`
	Total     float64 `json:"total"`
}

type quoteInput struct {
	Body QuoteRequest
}

type quoteOutput struct {
	Body Quote
}

// registerQuotes prices a request against the catalog without storing it.
// Rules the schema cannot express, like whether a product exists, are
// reported in the same shape as Huma's own validation errors.
func registerQuotes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "v1-create-quote",
		Method:      http.MethodPost,
		Path:        "/quotes",
		Summary:     "Quote an order",
		Description: "Validates the body against QuoteRequest and prices it from the catalog. Nothing is stored.",
		Tags:        []string{"v1"},
	}, func(_ context.Context, in *quoteInput) (*quoteOutput, error) {
		var problems []error
		quote := Quote{Currency: "EUR", Customer: in.Body.Customer, Lines: make([]QuoteItem, 0, len(in.Body.Lines))}
		quote.Customer.Email = strings.ToLower(quote.Customer.Email)

		for i, line := range in.Body.Lines {
			p, ok := findProduct(line.ProductID)
			if !ok {
				problems = append(problems, &huma.ErrorDetail{
					Location: fmt.Sprintf("body.lines[%d].productId", i),
					Message:  "unknown product",
					Value:    line.ProductID,
				})
				continue
			}
			total := cents(p.Price * float64(line.Quantity))
			quote.Lines = append(quote.Lines, QuoteItem{ProductID: p.ID, Name: p.Name, Quantity: line.Quantity, UnitPrice: p.Price, Total: total})
			quote.Subtotal = cents(quote.Subtotal + total)
		}

		switch in.Body.Coupon {
		case "":
		case "SAVE10":
			quote.Discount = cents(quote.Subtotal * 0.10)
		default:
			problems = append(problems, &huma.ErrorDetail{Location: "body.coupon", Message: "unknown coupon", Value: in.Body.Coupon})
		}

		if len(problems) > 0 {
			return nil, huma.Error422UnprocessableEntity("quote cannot be priced", problems...)
		}

		switch in.Body.Shipping {
		case "express":
			quote.Shipping = 14.90
		default:
			if quote.Subtotal < 100 {
				quote.Shipping = 4.90
			}
		}
		quote.Total = cents(quote.Subtotal - quote.Discount + quote.Shipping)
		quote.ExpiresAt = time.Now().UTC().Add(24 * time.Hour)
		return &quoteOutput{Body: quote}, nil
	})
}
