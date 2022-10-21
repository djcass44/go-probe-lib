package probe

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Handler struct {
	payload Payload
	shutdownCallbacks []func()
	serverCallbacks []func()
	cLock *sync.Mutex
}

func NewHandler() *Handler {
	m := &Handler{
		payload: Payload{
			Component: Component{
				Ok:     true,
				Status: StatusUp,
				Detail: "",
			},
			Components: []Component{},
		},
		cLock: &sync.Mutex{},
	}
	go m.Listen()

	return m
}

// Listen waits for SIGTERM and indicates
// to an observer that new requests
// shouldn't be sent to this instance.
func (h *Handler) Listen() {
	log.Print("starting readiness listener")
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, os.Interrupt)
	_ = <-sigC
	log.Print("received interrupt from the system")
	// stop responding that we're alive
	// so we can cleanly shutdown
	h.payload.Ok = false
	h.payload.Status = StatusDown
	h.payload.Detail = "shutdown signal received"

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

// RegisterShutdownServer adds an http server (or similar)
// that needs to be shutdown when the application is interrupted.
func (h *Handler) RegisterShutdownServer(f ShutdownAble) {
	h.cLock.Lock()
	h.serverCallbacks = append(h.serverCallbacks, func() {
		log.Printf("server shutdown: %s", f.Shutdown(context.TODO()))
	})
	h.cLock.Unlock()
}

type ShutdownAble interface {
	Shutdown(context.Context) error
}
