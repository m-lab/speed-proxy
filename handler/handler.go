package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// Config contains the configuration for the token handler.
type Config struct {
	ProjectID        string
	SecretName       string
	TokenExchangeURL string
	AllowedOrigin    string
	SMClient         *secretmanager.Client
	HTTPClient       *http.Client
}

// Handler handles token requests.
type Handler struct {
	config Config
}

// tokenRequest is the request body for the token exchange service.
type tokenRequest struct {
	APIKey string `json:"api_key"`
}

// tokenResponse is the response body from the token exchange service.
type tokenResponse struct {
	Token string `json:"token"`
}

// New creates a new Handler with the given configuration.
func New(config Config) *Handler {
	return &Handler{config: config}
}

// Token handles requests for JWT tokens.
func (h *Handler) Token(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight.
	if r.Method == http.MethodOptions {
		h.setCORSHeaders(w)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Only allow GET requests.
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.setCORSHeaders(w)

	ctx := r.Context()

	// Get the API key from Secret Manager.
	apiKey, err := h.getAPIKey(ctx)
	if err != nil {
		log.Printf("Failed to get API key: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Exchange the API key for a JWT token.
	token, err := h.exchangeToken(ctx, apiKey)
	if err != nil {
		log.Printf("Failed to exchange token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return the token.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokenResponse{Token: token})
}

// setCORSHeaders sets the CORS headers for the response.
func (h *Handler) setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", h.config.AllowedOrigin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Max-Age", "86400")
}

// getAPIKey retrieves the API key from Secret Manager.
func (h *Handler) getAPIKey(ctx context.Context) (string, error) {
	name := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", h.config.ProjectID, h.config.SecretName)

	result, err := h.config.SMClient.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	})
	if err != nil {
		return "", fmt.Errorf("failed to access secret version: %w", err)
	}

	return string(result.Payload.Data), nil
}

// exchangeToken exchanges an API key for a JWT token.
func (h *Handler) exchangeToken(ctx context.Context, apiKey string) (string, error) {
	reqBody, err := json.Marshal(tokenRequest{APIKey: apiKey})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.config.TokenExchangeURL, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.config.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return tokenResp.Token, nil
}
