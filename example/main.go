package main

import (
	"flag"
	"fmt"
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
	router := http.NewServeMux()
	router.Handle("/", http.FileServer(http.FS(os.DirFS(*assetDir))))

	// start the http server in the
	// background
	go func() {
		addr := fmt.Sprintf(":%d", *port)
		log.Printf("starting server on interface %s", addr)
		log.Fatal(http.ListenAndServe(addr, router))
	}()
	// wait for a signal
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGTERM, syscall.SIGINT)
	sig := <-sigC
	log.Printf("received shutdown signal (%s)", sig)
}