package env

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	ratesAPIURLKey     string = "RATES_API_URL"
	ratesAPIURLDefault string = "https://api.frankfurter.dev/v1"
)

// RatesAPIURL is the exchange-rate API behind the example price route.
func RatesAPIURL() string {
	return lookup(ratesAPIURLKey, ratesAPIURLDefault, parseRatesAPIURL)
}

func parseRatesAPIURL(s string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(s))
	if err != nil {
		return "", err
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", fmt.Errorf("rates API URL %q must be an absolute http(s) URL", s)
	}
	return strings.TrimRight(u.String(), "/"), nil
}
