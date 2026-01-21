package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/m-lab/go/flagx"
	"github.com/m-lab/speed-proxy/handler"
)

var (
	listenAddr       = flag.String("listen-addr", ":8080", "Address to listen on")
	projectID        = flag.String("project-id", "", "GCP project ID for Secret Manager")
	secretName       = flag.String("secret-name", "", "Name of the secret containing the API key")
	tokenExchangeURL = flag.String("token-exchange-url", "https://auth.measurementlab.net/v0/token/integration", "URL of the token exchange service")
	allowedOrigin    = flag.String("allowed-origin", "https://speed.measurementlab.net", "Allowed CORS origin")
)

func main() {
	flag.Parse()
	flagx.ArgsFromEnv(flag.CommandLine)

	if *projectID == "" {
		log.Fatal("-project-id is required")
	}
	if *secretName == "" {
		log.Fatal("-secret-name is required")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize Secret Manager client.
	smClient, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create Secret Manager client: %v", err)
	}
	defer smClient.Close()

	// Create the token handler.
	h := handler.New(handler.Config{
		ProjectID:        *projectID,
		SecretName:       *secretName,
		TokenExchangeURL: *tokenExchangeURL,
		AllowedOrigin:    *allowedOrigin,
		SMClient:         smClient,
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
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for shutdown signal.
	<-ctx.Done()
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}
}
