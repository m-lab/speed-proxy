# speed-proxy

[![Test](https://github.com/m-lab/speed-proxy/actions/workflows/test.yml/badge.svg)](https://github.com/m-lab/speed-proxy/actions/workflows/test.yml)
[![Coverage Status](https://coveralls.io/repos/github/m-lab/speed-proxy/badge.svg?branch=main)](https://coveralls.io/github/m-lab/speed-proxy?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/m-lab/speed-proxy)](https://goreportcard.com/report/github.com/m-lab/speed-proxy)
[![Go Version](https://img.shields.io/github/go-mod/go-version/m-lab/speed-proxy)](https://go.dev/)
[![Go Reference](https://pkg.go.dev/badge/github.com/m-lab/speed-proxy.svg)](https://pkg.go.dev/github.com/m-lab/speed-proxy)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/m-lab/speed-proxy)

Integrator backend service for speed.measurementlab.net. This service acts as a
security boundary between the frontend client and M-Lab's token exchange
service.

## Overview

The service provides a single endpoint that:

1. Exchanges the M-Lab API key for a short-lived JWT token via M-Lab's token exchange service
2. Returns the JWT to the frontend client

The frontend then uses this JWT to access M-Lab's Locate API at
`/v2/priority/nearest`.

## Configuration

| Flag | Environment Variable | Default | Description |
|------|---------------------|---------|-------------|
| `-listen-addr` | `LISTEN_ADDR` | `:8080` | Address to listen on |
| `-api-key` | `API_KEY` | (required) | M-Lab API key for token exchange |
| `-token-exchange-url` | `TOKEN_EXCHANGE_URL` | `https://auth.mlab-sandbox.measurementlab.net/v0/token/integration` | URL of the token exchange service |
| `-allowed-origin` | `ALLOWED_ORIGIN` | `https://speed.measurementlab.net` | Allowed CORS origin |

## API

### GET /v0/token

Returns a JWT token for authenticating with M-Lab's Locate API.

**Response:**
```json
{
  "token": "<jwt-token>"
}
```

### GET /health

Health check endpoint. Returns `200 OK` with body `ok`.

## Deployment

### Deploy to Cloud Run

```bash
gcloud run deploy speed-proxy \
  --source . \
  --region us-central1 \
  --set-env-vars "API_KEY=mlabk.ki_xxx.secret" \
  --allow-unauthenticated
```

## Local Development

```bash
API_KEY="mlabk.ki_xxx.secret" go run . -allowed-origin="http://localhost:3000"
```

## Docker

```bash
# Build
docker build -t speed-proxy .

# Run
export API_KEY="mlabk.ki_xxx.secret"
docker run -p 8080:8080 -e API_KEY speed-proxy
```
