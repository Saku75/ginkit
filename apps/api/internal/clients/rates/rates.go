// Package rates reads exchange rates from a Frankfurter-compatible API
// (https://frankfurter.dev). It is the example of a client that
// cmd/api/main.go builds once and handlers share; replace it with the SDKs
// the application needs.
package rates

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// ErrUnknownCurrency reports a currency code the API does not serve.
var ErrUnknownCurrency = errors.New("unknown currency")

type Client struct {
	baseURL string
	http    *http.Client
}

// New takes the API's base URL without a trailing slash and the HTTP client to
// call it with, so timeouts and transports stay the caller's decision.
func New(baseURL string, httpClient *http.Client) *Client {
	return &Client{baseURL: baseURL, http: httpClient}
}

// Ping reports whether the API answers, for a readiness check.
func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/currencies", nil)
	if err != nil {
		return err
	}
	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("rates: %w", err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("rates: unexpected status %s", res.Status)
	}
	return nil
}

// Rate returns how many units of to one unit of from buys.
func (c *Client) Rate(ctx context.Context, from, to string) (float64, error) {
	if from == to {
		return 1, nil
	}

	query := url.Values{"base": {from}, "symbols": {to}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/latest?"+query.Encode(), nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return 0, fmt.Errorf("rates: %w", err)
	}
	defer res.Body.Close()

	switch {
	case res.StatusCode == http.StatusNotFound:
		return 0, ErrUnknownCurrency
	case res.StatusCode != http.StatusOK:
		return 0, fmt.Errorf("rates: unexpected status %s", res.Status)
	}

	var body struct {
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&body); err != nil {
		return 0, fmt.Errorf("rates: decode response: %w", err)
	}

	rate, ok := body.Rates[to]
	if !ok {
		return 0, ErrUnknownCurrency
	}
	return rate, nil
}
