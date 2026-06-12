# syntax=docker/dockerfile:1

# --- Build stage ---
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=v1.4.0
RUN COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown") && \
    DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') && \
    CGO_ENABLED=0 go build \
      -ldflags "-X github.com/axiomod/axiomod/framework/version.Version=${VERSION} \
                -X github.com/axiomod/axiomod/framework/version.GitCommit=${COMMIT} \
                -X github.com/axiomod/axiomod/framework/version.BuildDate=${DATE}" \
      -o /out/axiomod-server ./cmd/axiomod-server

# --- Runtime stage ---
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -u 10001 axiomod

WORKDIR /app

COPY --from=builder /out/axiomod-server /app/axiomod-server
COPY configs/ /app/configs/

USER axiomod

# HTTP, gRPC, metrics
EXPOSE 8080 9090 9091

ENTRYPOINT ["/app/axiomod-server"]
