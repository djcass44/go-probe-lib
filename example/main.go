/*
 *    Copyright 2022 Django Cass
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 *
 */

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/djcass44/go-probe-lib/pkg/probe"
	"github.com/djcass44/go-utils/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	port := flag.Int("port", 8080, "http port to run on (default: 8080)")
	healthPort := flag.Int("health-port", 8081, "http port for health checks to run on (default: 8081)")
	assetDir := flag.String("asset-dir", "./assets", "static asset folder to serve (default: ./assets)")

	flag.Parse()

	probes := probe.NewHandler(time.Second)
	zc := zap.NewProductionConfig()
	zc.Level = zap.NewAtomicLevelAt(zapcore.Level(-10))

	log, ctx := logging.NewZap(context.TODO(), zc)

	// configure routing
	fs := http.FileServer(http.FS(os.DirFS(*assetDir)))
	router := http.NewServeMux()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Info(fmt.Sprintf("%s %s %s", r.Method, r.URL.Path, r.UserAgent()))
		fs.ServeHTTP(w, r)
	})

	addr := fmt.Sprintf(":%d", *port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: time.Second * 3,
	}
	// register shutdown functions
	probes.RegisterShutdownServer(ctx, srv)
	probes.RegisterShutdownFunc(func() {
		log.Info("I'm a slow shutdown func!!")
		time.Sleep(time.Second * 10)
		log.Info("cya!")
	})
	go func() {
		if err := probes.ListenAndServe(ctx, *healthPort); err != nil {
			os.Exit(1)
		}
	}()

	// start the http server in the
	// background
	go func() {
		log.Info("starting server", "Interface", addr)
		log.Error(srv.ListenAndServe(), "http server exited")
	}()
	// wait for a signal
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, os.Kill)
	sig := <-sigC
	log.Info("received shutdown signal", "Signal", sig)
}
