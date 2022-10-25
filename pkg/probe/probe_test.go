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
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandler_timeout(t *testing.T) {
	ctx := logr.NewContext(context.TODO(), testr.NewWithOptions(t, testr.Options{Verbosity: 10}))
	port := rand.Intn(30000) + 2000
	h := NewHandler(time.Second * 1)
	go func() {
		_ = h.ListenAndServe(ctx, port)
	}()
	assert.ErrorIs(t, h.onShutdown(ctx), ErrDeadlineExceeded)
}

func TestHandler_RegisterShutdownFunc(t *testing.T) {
	ctx := logr.NewContext(context.TODO(), testr.NewWithOptions(t, testr.Options{Verbosity: 10}))
	called := false
	port := rand.Intn(30000) + 2000
	h := NewHandler(0)
	h.RegisterShutdownFunc(func() {
		called = true
	})
	// start the server
	go func() {
		_ = h.ListenAndServe(ctx, port)
	}()

	assert.NoError(t, h.onShutdown(ctx))

	time.Sleep(time.Second)
	assert.True(t, called)
}

func TestHandler_RegisterShutdownServer(t *testing.T) {
	ctx := logr.NewContext(context.TODO(), testr.NewWithOptions(t, testr.Options{Verbosity: 10}))
	ts := httptest.NewServer(http.NotFoundHandler())
	ts.Config.RegisterOnShutdown(func() {
		t.Log("shutting down test server")
	})
	defer ts.Close()

	port := rand.Intn(30000) + 2000
	h := NewHandler(0)
	h.RegisterShutdownServer(ctx, ts.Config)
	// start the server
	go func() {
		_ = h.ListenAndServe(ctx, port)
	}()

	// assert that the server is running
	_, err := ts.Client().Get(ts.URL)
	assert.NoError(t, err)

	assert.NoError(t, h.onShutdown(ctx))

	time.Sleep(time.Millisecond * 100)
	_, err = ts.Client().Get(ts.URL)
	t.Log(err)
	assert.Error(t, err)
}
