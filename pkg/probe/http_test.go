package probe

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_Livez(t *testing.T) {
	t.Run("dead handler returns non-ok status code", func(t *testing.T) {
		h := new(Handler)
		h.isDead = true
		req := httptest.NewRequest(http.MethodGet, "https://example.org", nil)
		w := httptest.NewRecorder()

		h.Livez(w, req)
		assert.EqualValues(t, http.StatusServiceUnavailable, w.Code)
	})
	t.Run("alive handler returns ok status code", func(t *testing.T) {
		h := new(Handler)
		req := httptest.NewRequest(http.MethodGet, "https://example.org", nil)
		w := httptest.NewRecorder()

		h.Livez(w, req)
		assert.EqualValues(t, http.StatusOK, w.Code)
	})
}

func TestHandler_writeJSON(t *testing.T) {
	h := new(Handler)

	var cases = []struct {
		name    string
		payload any
		out     string
	}{
		{
			"payloads that cant be marshalled return an error",
			make(chan struct{}),
			messageGenericError,
		},
		{
			"standard objects can be marshalled",
			struct {
				Status string `json:"status"`
			}{"up"},
			`{"status": "up"}`,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			h.writeJSON(w, tt.payload, true)

			resp := w.Body.String()
			assert.JSONEq(t, tt.out, resp)
		})
	}
}
