FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o speed-proxy .

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/speed-proxy /speed-proxy

ENTRYPOINT ["/speed-proxy"]
