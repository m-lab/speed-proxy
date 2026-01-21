# speed-proxy

Integrator backend service for speed.measurementlab.net. This service acts as a
security boundary between the frontend client and M-Lab's token exchange
service.

## Overview

The service provides a single endpoint that:

1. Retrieves the M-Lab API key from Google Secret Manager
2. Exchanges the API key for a short-lived JWT token via M-Lab's token exchange service
3. Returns the JWT to the frontend client

The frontend then uses this JWT to access M-Lab's Locate API at
`/v2/priority/nearest`.

## Configuration

| Flag | Environment Variable | Default | Description |
|------|---------------------|---------|-------------|
| `-listen-addr` | `LISTEN_ADDR` | `:8080` | Address to listen on |
| `-project-id` | `PROJECT_ID` | (required) | GCP project ID for Secret Manager |
| `-secret-name` | `SECRET_NAME` | (required) | Name of the secret containing the API key |
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

### Prerequisites

1. Create a secret in Secret Manager containing the M-Lab API key:
   ```bash
   echo -n "mlabk.ki_xxx.secret" | gcloud secrets create mlab-api-key \
     --data-file=- \
     --project=YOUR_PROJECT_ID
   ```

2. Grant the Cloud Run service account access to the secret:
   ```bash
   gcloud secrets add-iam-policy-binding mlab-api-key \
     --member="serviceAccount:YOUR_SERVICE_ACCOUNT" \
     --role="roles/secretmanager.secretAccessor" \
     --project=YOUR_PROJECT_ID
   ```

### Deploy to Cloud Run

```bash
gcloud run deploy speed-proxy \
  --source . \
  --region us-central1 \
  --set-env-vars "PROJECT_ID=YOUR_PROJECT_ID,SECRET_NAME=mlab-api-key" \
  --allow-unauthenticated
```

## Local Development

```bash
# Set up Application Default Credentials
gcloud auth application-default login

# Run locally
go run . \
  -project-id=YOUR_PROJECT_ID \
  -secret-name=mlab-api-key \
  -allowed-origin="http://localhost:3000"
```
