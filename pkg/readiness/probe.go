package readiness

import (
	"github.com/djcass44/go-probe-lib/pkg/probe"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Manager struct {
	probe.HttpHandler
}

func NewManager() *Manager {
	m := &Manager{
		HttpHandler: probe.HttpHandler{
			Payload: probe.Payload{
				Component: probe.Component{
					Ok:     true,
					Status: probe.StatusUp,
					Detail: "",
				},
				Components: []probe.Component{},
			},
		},
	}
	go m.Listen()

	return m
}

// Listen waits for SIGTERM and indicates
// to an observer that new requests
// shouldn't be sent to this instance.
func (m *Manager) Listen() {
	log.Print("starting readiness listener")
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGTERM)
	_ = <-sigC
	log.Print("received SIGTERM")
	// stop responding that we're alive
	// so we can cleanly shutdown
	m.Payload.Ok = false
	m.Payload.Status = probe.StatusDown
	m.Payload.Detail = "shutdown signal received"
}
