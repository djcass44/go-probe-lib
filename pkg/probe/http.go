package probe

import (
	"encoding/json"
	"log"
	"net/http"
)

func (h *HttpHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// convert the data to JSON
	data, err := json.Marshal(h.Payload)
	if err != nil {
		// if we couldn't convert it, then
		// manually write a JSON message
		log.Printf("error: failed to convert response to json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"status": "internal server error", "ok": false}`))
		return
	}
	// if we're not ok, tell the observer (probably k8s)
	// that we shouldn't receive new traffic
	if !h.Payload.Ok {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	_, _ = w.Write(data)
}
