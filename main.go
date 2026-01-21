package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/m-lab/go/flagx"
	"github.com/m-lab/go/rtx"
	"github.com/m-lab/speed-proxy/handler"
)

var (
	listenAddr       = flag.String("listen-addr", ":8080", "Address to listen on")
	apiKey           = flag.String("api-key", "", "API key for token exchange")
	tokenExchangeURL = flag.String("token-exchange-url", "https://auth.mlab-sandbox.measurementlab.net/v0/token/integration", "URL of the token exchange service")
	allowedOrigin    = flag.String("allowed-origin", "https://speed.measurementlab.net", "Allowed CORS origin")
)

func main() {
	flag.Parse()
	flagx.ArgsFromEnv(flag.CommandLine)

	if *apiKey == "" {
		log.Fatal("-api-key is required")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Create the token handler.
	h := handler.New(handler.Config{
		APIKey:           *apiKey,
		TokenExchangeURL: *tokenExchangeURL,
		AllowedOrigin:    *allowedOrigin,
		HTTPClient:       &http.Client{Timeout: 10 * time.Second},
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/v0/token", h.Token)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:         *listenAddr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine.
	go func() {
		log.Printf("Starting server on %s", *listenAddr)
		err := server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			rtx.Must(err, "Server error")
		}
	}()

	// Wait for shutdown signal.
	<-ctx.Done()
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	rtx.Must(server.Shutdown(shutdownCtx), "Server shutdown error")
}
