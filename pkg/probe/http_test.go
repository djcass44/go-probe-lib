/*
 *    Copyright 2022 Django Cass
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 *
 */

package probe

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
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
	ctx := logr.NewContext(context.TODO(), testr.NewWithOptions(t, testr.Options{Verbosity: 10}))
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
			h.writeJSON(ctx, w, tt.payload, true)

			resp := w.Body.String()
			assert.JSONEq(t, tt.out, resp)
		})
	}
}
