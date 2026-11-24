package v1

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"ginkit/api/internal/clients/rates"
)

type getPriceInput struct {
	ID       string `path:"id" pattern:"^p-[0-9]+$" example:"p-1001"`
	Currency string `query:"currency" default:"EUR" pattern:"^[A-Z]{3}$" doc:"ISO 4217 code to quote the price in."`
}

type getPriceOutput struct {
	Body struct {
		ProductID string  `json:"productId" example:"p-1001"`
		Currency  string  `json:"currency" example:"USD"`
		Amount    float64 `json:"amount" example:"97.36"`
		Rate      float64 `json:"rate" example:"1.094" doc:"EUR to currency rate applied."`
	}
}

// registerPrices is the handler that calls an injected client. Errors the
// client reports as the caller's fault become 422 with the field named; all
// others are logged and answered with 502, without leaking what went wrong.
func registerPrices(api huma.API, fx Rates) {
	huma.Register(api, huma.Operation{
		OperationID: "v1-get-product-price",
		Method:      http.MethodGet,
		Path:        "/products/{id}/price",
		Summary:     "Price in a currency",
		Description: "Converts the EUR price with a rate from the exchange-rate client that main builds once.",
		Tags:        []string{"v1"},
		Errors:      []int{http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusBadGateway},
	}, func(ctx context.Context, in *getPriceInput) (*getPriceOutput, error) {
		p, ok := findProduct(in.ID)
		if !ok {
			return nil, huma.Error404NotFound(fmt.Sprintf("product %s not found", in.ID))
		}

		rate, err := fx.Rate(ctx, "EUR", in.Currency)
		switch {
		case errors.Is(err, rates.ErrUnknownCurrency):
			return nil, huma.Error422UnprocessableEntity("unknown currency", &huma.ErrorDetail{
				Location: "query.currency",
				Message:  "not a currency the exchange-rate API serves",
				Value:    in.Currency,
			})
		case err != nil:
			slog.ErrorContext(ctx, "fetching exchange rate", slog.Any("error", err))
			return nil, huma.Error502BadGateway("exchange rates are unavailable")
		}

		out := &getPriceOutput{}
		out.Body.ProductID = p.ID
		out.Body.Currency = in.Currency
		out.Body.Rate = rate
		out.Body.Amount = cents(p.Price * rate)
		return out, nil
	})
}

func cents(v float64) float64 {
	return math.Round(v*100) / 100
}
