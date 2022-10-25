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
