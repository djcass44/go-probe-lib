package probe

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandler_timeout(t *testing.T) {
	port := rand.Intn(30000) + 2000
	h := NewHandler(time.Second * 1)
	go func() {
		_ = h.ListenAndServe(port)
	}()
	assert.ErrorIs(t, h.onShutdown(), ErrDeadlineExceeded)
}

func TestHandler_RegisterShutdownFunc(t *testing.T) {
	called := false
	port := rand.Intn(30000) + 2000
	h := NewHandler(0)
	h.RegisterShutdownFunc(func() {
		called = true
	})
	// start the server
	go func() {
		_ = h.ListenAndServe(port)
	}()

	assert.NoError(t, h.onShutdown())

	time.Sleep(time.Second)
	assert.True(t, called)
}

func TestHandler_RegisterShutdownServer(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	ts.Config.RegisterOnShutdown(func() {
		t.Log("shutting down test server")
	})
	defer ts.Close()

	port := rand.Intn(30000) + 2000
	h := NewHandler(0)
	h.RegisterShutdownServer(ts.Config)
	// start the server
	go func() {
		_ = h.ListenAndServe(port)
	}()

	// assert that the server is running
	_, err := ts.Client().Get(ts.URL)
	assert.NoError(t, err)

	assert.NoError(t, h.onShutdown())

	time.Sleep(time.Millisecond * 100)
	_, err = ts.Client().Get(ts.URL)
	t.Log(err)
	assert.Error(t, err)
}
