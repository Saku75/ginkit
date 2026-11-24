package handlers

import "ginkit/api/internal/clients/rates"

// Dependencies carries the clients handlers call. cmd/api/main.go builds each
// one once and sets it here; Register hands every API version only what it
// uses, as an interface that version declares. Add a field per client.
type Dependencies struct {
	Rates *rates.Client
}
