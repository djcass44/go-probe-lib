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
	"encoding/json"
	"github.com/go-logr/logr"
	"net/http"
)

func (h *Handler) Livez(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(r.Context(), w, &Component{Ok: !h.isDead}, !h.isDead)
}

func (h *Handler) Readyz(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(r.Context(), w, h.payload, h.payload.Ok)
}

func (*Handler) writeJSON(ctx context.Context, w http.ResponseWriter, payload any, ok bool) {
	log := logr.FromContextOrDiscard(ctx)
	w.Header().Set("Content-Type", "application/json")
	// convert the data to JSON
	data, err := json.Marshal(payload)
	if err != nil {
		// if we couldn't convert it, then
		// manually write a JSON message
		log.Error(err, "failed to convert response to json")
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
