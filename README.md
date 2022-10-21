# Go Probe Lib

Go library for writing proper liveness/readiness probes.
It's designed for use in a Kubernetes environment, however it is not bound to it.

## How it works

This library is simple in design and uses the following flow:

1. Allow the caller to register shutdown functions to gracefully stop the application
2. Wait for an interrupt signal from the observer (e.g. Kubelet)
3. Mark the `/readyz` endpoint as failed (so that Kubernetes removes the Pod from service)
4. Call all shutdown functions (http servers are shutdown last)
5. Mark the `/livez` endpoint as failed (so that Kubernetes kills the Pod)
6. Wait for the observer to send a kill signal or optionally exit after a specified duration

## Usage

For a complete example, take a look at the [sample app](./example/main.go).

```commandline
go get github.com/djcass44/go-probe-lib
```

```go
package main

import (
	"context"
	"github.com/djcass44/go-probe-lib/pkg/probe"
	"log"
	"os"
	"time"
)

func main() {
	// set a 30s deadline for killing the application
	// if the kubelet hasn't done it already
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
		if err := probes.ListenAndServe(context.TODO(), 8081);
			err != nil {
			os.Exit(1)
		}
	}()
}
```

### Logging

This project uses the `go-logr` logging facade so that you can provide your own logger.
To ensure logs are emitted correctly, make sure that the `context.Context` you provide contains your `LogSink`.

```go
// this bit will depend on the logging
// solution you choose to use
log := foobar.NewLogger()
ctx := logr.NewContext(context.TODO(), log)
...
probes.ListenAndServe(ctx, 8081)
```
