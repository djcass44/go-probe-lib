# Go Probe Lib

Go library for writing proper liveness/readiness probes.

## How it works

This library is simple in design and uses the following flow:

1. Allow the caller to register shutdown functions
2. Wait for a `SIGTERM` signal from the OS
3. Mark the `/readyz` endpoint as failed
4. Call all shutdown functions (http servers are shutdown last)
5. Mark the `/livez` endpoint as failed
6. Wait for the observer (e.g. Kubelet) to send a `SIGKILL`

## Usage

For a complete example, take a look at the [sample app](./example/main.go).

```commandline
go get github.com/djcass44/go-probe-lib
```

```go
package main

import (
	"github.com/djcass44/go-probe-lib/pkg/probe"
	"os"
	"time"
)

func main() {
	probes := probe.NewHandler(time.Second * 30)
	// register one or more functions
	probes.RegisterShutdownFunc(func() {
		log.Print("I'm a slow shutdown func!!")
		time.Sleep(time.Second * 10)
		log.Print("cya!")
	})
	// don't forget to call this in a goroutine
	// otherwise it will block.
	go func() {
		if err := probes.ListenAndServe(8081); err != nil {
			os.Exit(1)
		}
	}()
}
```
