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
