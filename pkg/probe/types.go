package probe

type HttpHandler struct {
	Payload Payload
}

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
