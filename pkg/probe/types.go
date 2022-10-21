package probe

import "errors"

type Payload struct {
	Component
	Components []Component `json:"components"`
}

type Component struct {
	Ok     bool   `json:"ok"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

const (
	StatusUp      = "up"
	StatusDown    = "down"
	StatusUnknown = "unknown"
)

const messageGenericError = `{"status": "internal server error", "ok": false}`

var ErrDeadlineExceeded = errors.New("observer-initiated shutdown deadline exceeded")
