# syntax=docker/dockerfile:1
FROM golang:1.26.1-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o complianced ./cmd/complianced/

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/complianced /usr/local/bin/complianced
EXPOSE 8091

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8091/healthz || exit 1

ENTRYPOINT ["complianced"]
