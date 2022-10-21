package main

import (
	"flag"
	"fmt"
	"github.com/djcass44/go-probe-lib/pkg/probe"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	port := flag.Int("port", 8080, "http port to run on (default: 8080)")
	assetDir := flag.String("asset-dir", "./assets", "static asset folder to serve")

	flag.Parse()

	probes := probe.NewHandler()

	// configure routing
	fs := http.FileServer(http.FS(os.DirFS(*assetDir)))
	router := http.NewServeMux()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.UserAgent())
		fs.ServeHTTP(w, r)
	})
	// add the probes (this is the important bit)
	router.HandleFunc("/healthz/readyz", probes.Readyz)
	router.HandleFunc("/healthz/livez", probes.Livez)

	addr := fmt.Sprintf(":%d", *port)
	srv := &http.Server{
		Addr: addr,
		Handler: router,
	}
	// register shutdown functions
	probes.RegisterShutdownServer(srv)
	probes.RegisterShutdownFunc(func() {
		log.Print("I'm a slow shutdown func!!")
		time.Sleep(time.Second * 10)
		log.Print("cya!")
	})

	// start the http server in the
	// background
	go func() {
		log.Printf("starting server on interface %s", addr)
		log.Printf("error: http server exited: %s", srv.ListenAndServe())
	}()
	// wait for a signal
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, os.Kill)
	sig := <-sigC
	log.Printf("received shutdown signal (%s)", sig)
}