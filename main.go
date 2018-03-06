package main

//go:generate protoc -I $GOPATH/src/github.com/telecom-tower/towerapi/v1 telecomtower.proto --go_out=plugins=grpc:$GOPATH/src/github.com/telecom-tower/towerapi/v1
//go:generate esc -prefix html -ignore .DS_Store -o html.go -pkg main html

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/trace"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	ws2811 "github.com/supcik/web_ws281x_go"
	"github.com/telecom-tower/grpc-renderer"
)

func main() {
	debug := flag.Bool("debug", false, "Debug mode")
	verbose := flag.Bool("verbose", false, "Verbose mode")
	httpAddress := flag.String("http-address", "127.0.0.1", "listening HTTP address")
	traceFile := flag.String("trace", "", "Generate tracing file")
	httpPort := flag.Int("http-port", 8080, "listening HTTP port")
	grpcPort := flag.Int("grpc-port", 10000, "listening gRPC port")

	flag.Parse()
	if *debug {
		log.SetLevel(log.DebugLevel)
	} else if *verbose {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}

	if *traceFile != "" {
		f, err := os.Create(*traceFile)
		if err != nil {
			log.Panic(errors.WithMessage(err, "Unable to create trace file"))
		}
		err = trace.Start(f)
		defer trace.Stop()
		if err != nil {
			log.Panic(errors.WithMessage(err, "Unable to trace"))
		}
	}

	// Create and run hub
	hub := ws2811.NewHub()
	wsopt := ws2811.DefaultOptions
	wsopt.Channels[0].LedCount = 1024
	ws, err := ws2811.MakeWS2811(&wsopt, hub)
	if err != nil {
		log.Fatal(err)
	}
	if err = ws.Init(); err != nil {
		log.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(3)
	go func() {
		hub.Run()
		log.Fatalf("Hub is no longer running")
		wg.Done()
	}()

	r := mux.NewRouter()
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Serving websocket")
		ws2811.ServeWs(hub, w, r)
	})
	r.PathPrefix("/").Handler(http.FileServer(FS(false))) // nolint
	// r.PathPrefix("/").Handler(http.FileServer(http.Dir("html"))) // nolint

	grpcLis, err := net.Listen("tcp", fmt.Sprintf(":%d", *grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	go func() {
		log.Fatal(renderer.Serve(grpcLis, ws))
		wg.Done()
	}()

	a := fmt.Sprintf("%s:%d", *httpAddress, *httpPort)
	srv := &http.Server{
		Handler: r,
		Addr:    a,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Infof("HTTP server ready at http://%s", a)
	go func() {
		log.Fatal(srv.ListenAndServe())
		wg.Done()
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			trace.Stop()
			log.Info("Finished")
			os.Exit(0)
		}
	}()

	wg.Wait()
}
