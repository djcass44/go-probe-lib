package probe

import (
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

		// todo replace with assertions
		if w.Code/100 != 5 {
			t.Fail()
		}
	})
	t.Run("alive handler returns ok status code", func(t *testing.T) {
		h := new(Handler)
		req := httptest.NewRequest(http.MethodGet, "https://example.org", nil)
		w := httptest.NewRecorder()

		h.Livez(w, req)

		// todo replace with assertions
		if w.Code != http.StatusOK {
			t.Fail()
		}
	})
}
