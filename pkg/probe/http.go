package probe

import (
	"encoding/json"
	"log"
	"net/http"
)

func (h *Handler) Livez(w http.ResponseWriter, _ *http.Request) {
	h.writeJSON(w, &Component{Ok: !h.isDead}, !h.isDead)
}

func (h *Handler) Readyz(w http.ResponseWriter, _ *http.Request) {
	h.writeJSON(w, h.payload, h.payload.Ok)
}

func (*Handler) writeJSON(w http.ResponseWriter, payload any, ok bool) {
	w.Header().Set("Content-Type", "application/json")
	// convert the data to JSON
	data, err := json.Marshal(payload)
	if err != nil {
		// if we couldn't convert it, then
		// manually write a JSON message
		log.Printf("error: failed to convert response to json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(messageGenericError))
		return
	}
	// if we're not ok, tell the observer (probably k8s)
	// that we shouldn't receive new traffic
	if !ok {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	_, _ = w.Write(data)
}
