package probe

import (
	"context"
	"fmt"
	"log"
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
// when the application exits.
func (h *Handler) ListenAndServe(port int) error {
	// start the http server in the background
	// on the user-specified port
	go func() {
		router := http.NewServeMux()
		router.HandleFunc("/livez", h.Livez)
		router.HandleFunc("/readyz", h.Readyz)
		addr := fmt.Sprintf(":%d", port)
		log.Printf("starting health server on interface %s", addr)
		if err := http.ListenAndServe(addr, router); err != nil {
			log.Printf("error: healthz server exited: %s", err)
			return
		}
	}()
	return h.Listen()
}

// Listen waits for SIGTERM and indicates
// to an observer that new requests
// shouldn't be sent to this instance.
func (h *Handler) Listen() error {
	log.Print("starting readiness listener")
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, os.Interrupt)
	_ = <-sigC
	log.Print("received interrupt from the system")
	return h.onShutdown()
}

// onShutdown is the series of actions taken
// as we initiate shutdown. It's a separate
// function so that we can privately test it.
func (h *Handler) onShutdown() error {
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
		log.Printf("starting shutdown hook %d/%d (%s elapsed)", i+1, numCallbacks, time.Since(start))
		f()
		log.Printf("finished shutdown hook %d/%d (%s elapsed - %s total)", i+1, numCallbacks, time.Since(s2), time.Since(start))
	}
	h.isDead = true
	if h.killDuration > 0 {
		time.Sleep(h.killDuration)
		log.Printf("exiting due to shutdown timeout (%s)", h.killDuration)
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

// RegisterShutdownServer adds n http server (or similar)
// that needs to be shutdown when the application is interrupted.
func (h *Handler) RegisterShutdownServer(f ShutdownAble) {
	h.cLock.Lock()
	h.serverCallbacks = append(h.serverCallbacks, func() {
		// ask the server to shut down
		if err := f.Shutdown(context.TODO()); err != nil {
			log.Printf("error: shutting down server: %s", err)
		}
	})
	h.cLock.Unlock()
}

type ShutdownAble interface {
	Shutdown(context.Context) error
}
