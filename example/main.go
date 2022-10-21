package main

import (
	"flag"
	"fmt"
	"github.com/djcass44/go-probe-lib/pkg/readiness"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	port := flag.Int("port", 8080, "http port to run on (default: 8080)")
	assetDir := flag.String("asset-dir", "./assets", "static asset folder to serve")

	flag.Parse()

	// configure routing
	fs := http.FileServer(http.FS(os.DirFS(*assetDir)))
	router := http.NewServeMux()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.UserAgent())
		fs.ServeHTTP(w, r)
	})
	// add the probes (this is the important bit)
	router.Handle("/healthz/readyz", readiness.NewManager())

	// start the http server in the
	// background
	go func() {
		addr := fmt.Sprintf(":%d", *port)
		log.Printf("starting server on interface %s", addr)
		log.Fatal(http.ListenAndServe(addr, router))
	}()
	// wait for a signal
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT)
	sig := <-sigC
	log.Printf("received shutdown signal (%s)", sig)
}