package probe

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Handler struct {
	payload           Payload
	shutdownCallbacks []func()
	serverCallbacks   []func()
	cLock             *sync.Mutex

	// isShutdown indicates that we're in the process
	// of shutting down the application.
	isShutdown bool
	// isDead indicates that the application is fully
	// shutdown and should be killed by the observer.
	isDead bool

	// killDuration is the time after which
	// we should exit the application if the
	// observer doesn't do it for us.
	// 0 means we won't exit until the
	// observer sends the kill signal
	killDuration time.Duration
}

// NewHandler creates a new instance of Handler
// with a specified timeout. A timeout of 0
// means that the application will not exit
// until the observer sends a kill signal.
//
// The handler requires starting with ListenAndServe
func NewHandler(timeout time.Duration) *Handler {
	m := &Handler{
		payload: Payload{
			Component: Component{
				Ok:     true,
				Status: StatusUp,
				Detail: "",
			},
			Components: []Component{},
		},
		cLock:        &sync.Mutex{},
		killDuration: timeout,
	}

	return m
}

// ListenAndServe starts a http server
// that serves HTTP requests from an observer
// (e.g. the kubelet).
//
// This server should only be used for health
// information and is only shutdown
// when the application exits (i.e. not gracefully).
//
// Returns an error if the specified deadline has exceeded
// and indicates that the caller (you) should exit the application
// using os.Exit
func (h *Handler) ListenAndServe(ctx context.Context, port int) error {
	log := logr.FromContextOrDiscard(ctx)
	// start the http server in the background
	// on the user-specified port
	go func() {
		router := http.NewServeMux()
		router.HandleFunc("/livez", h.Livez)
		router.HandleFunc("/readyz", h.Readyz)
		addr := fmt.Sprintf(":%d", port)
		log.V(2).Info("starting healthz server", "Interface", addr)
		if err := http.ListenAndServe(addr, router); err != nil {
			log.Error(err, "healthz server exited")
			return
		}
	}()
	return h.Listen(ctx)
}

// Listen waits for SIGTERM and indicates
// to an observer that new requests
// shouldn't be sent to this instance.
//
// Only call this if you're managing the HTTP
// components yourself. Generally you should
// just use ListenAndServe
func (h *Handler) Listen(ctx context.Context) error {
	log := logr.FromContextOrDiscard(ctx)
	log.V(2).Info("starting readiness listener")
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, os.Interrupt)
	sig := <-sigC
	log.V(1).Info("received interrupt from the system", "Signal", sig)
	log.Info("shutting down due to receiving system interrupt")
	return h.onShutdown(ctx)
}

// onShutdown is the series of actions taken
// as we initiate shutdown. It's a separate
// function so that we can privately test it.
func (h *Handler) onShutdown(ctx context.Context) error {
	log := logr.FromContextOrDiscard(ctx)
	// stop responding that we're alive, so
	// we can cleanly shut down
	h.payload.Ok = false
	h.payload.Status = StatusDown
	h.payload.Detail = "shutdown signal received"
	h.isShutdown = true

	// call all of our callbacks
	callbacks := append(h.shutdownCallbacks, h.serverCallbacks...)
	numCallbacks := len(callbacks)
	start := time.Now()
	for i, f := range callbacks {
		s2 := time.Now()
		log.V(4).Info("starting shutdown hook", "Current", i+1, "Total", numCallbacks, "Elapsed", time.Since(start))
		f()
		log.V(4).Info("finished shutdown hook", "Current", i+1, "Total", numCallbacks, "Elapsed", time.Since(s2), "TotalElapsed", time.Since(start))
	}
	h.isDead = true
	// if requested, wait for the deadline
	// and return an error.
	if h.killDuration > 0 {
		time.Sleep(h.killDuration)
		log.V(3).Info("exiting due to shutdown timeout", "Duration", h.killDuration)
		return ErrDeadlineExceeded
	}
	return nil
}

// RegisterShutdownFunc adds a user-provided function
// that needs to be called when the application is
// interrupted.
//
// Make sure to use RegisterShutdownServer for HTTP servers
// and http.Server RegisterOnShutdown for http-related shutdown
// callbacks.
func (h *Handler) RegisterShutdownFunc(f func()) {
	h.cLock.Lock()
	h.shutdownCallbacks = append(h.shutdownCallbacks, f)
	h.cLock.Unlock()
}

// RegisterShutdownServer adds a http server (or similar)
// that needs to be shutdown when the application is interrupted.
func (h *Handler) RegisterShutdownServer(ctx context.Context, f ShutdownAble) {
	log := logr.FromContextOrDiscard(ctx)
	h.cLock.Lock()
	h.serverCallbacks = append(h.serverCallbacks, func() {
		// ask the server to shut down
		if err := f.Shutdown(ctx); err != nil {
			log.V(4).Error(err, "shutting down server")
		}
	})
	h.cLock.Unlock()
}

// ShutdownAble describes any struct that implements
// the Shutdown function that http.Server uses.
//
// It only exists to ensure we're not bound to the
// net/http implementation of the HTTP server.
type ShutdownAble interface {
	Shutdown(context.Context) error
}
