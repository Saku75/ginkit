package env

import (
	"fmt"
	"regexp"
)

const (
	clientIPHeaderKey     string = "CLIENT_IP_HEADER"
	clientIPHeaderDefault string = ""
)

var headerNamePattern = regexp.MustCompile(`^[A-Za-z0-9-]+$`)

// ClientIPHeader names a header that always carries the real client IP, such
// as CF-Connecting-IP behind Cloudflare. It is trusted from any peer, so it is
// only safe when every request arrives through that platform.
func ClientIPHeader() string {
	return lookup(clientIPHeaderKey, clientIPHeaderDefault, parseClientIPHeader)
}

func parseClientIPHeader(s string) (string, error) {
	if !headerNamePattern.MatchString(s) {
		return "", fmt.Errorf("client IP header %q is not a valid header name", s)
	}
	return s, nil
}
