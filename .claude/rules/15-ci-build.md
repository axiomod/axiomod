# CI/CD and Build

## Build Commands

```
make build          # bin/axiomod-server (with ldflags: version, commit, date)
make build-cli      # bin/axiomod
make test           # go test -v ./...
make lint           # golangci-lint run ./...
make fmt            # go fmt ./...
make deps           # go mod tidy + download + install golangci-lint
make docker         # docker build -t axiomod/server:latest .
make clean          # go clean && rm -rf bin/
```

## Build Variables (injected via ldflags)

- `framework/version.Version` = v1.4.0
- `framework/version.GitCommit` = git rev-parse HEAD
- `framework/version.BuildDate` = UTC timestamp

## CI Pipeline (GitHub Actions)

**Triggers**: Push and PR to `main`/`master`.

**Steps**:
1. `go mod verify`
2. `gofmt -l .` (fail if unformatted)
3. `go vet ./...`
4. `go test -v -race ./...` (with race detector)
5. `go build -v ./cmd/axiomod-server`

**CodeQL**: Security analysis runs on push/PR + weekly cron.

## Pre-Submit Checklist

Before pushing:

```bash
make fmt             # Format code
make lint            # Run linter
make test            # Run tests
axiomod validator architecture  # Check import rules
```

## Linting

No `.golangci.yml` -- runs with default golangci-lint settings (errcheck, gosimple, govet, ineffassign, staticcheck, unused).

## Coverage Target

- **>80%** for core framework modules
- CI uses `-race` flag (Makefile does not -- run `go test -race ./...` locally too)

## Docker

```yaml
# docker-compose.reference.yaml infrastructure:
# PostgreSQL 14, Redis 6, Kafka/Zookeeper, Jaeger, Prometheus
# App: ports 8080 (HTTP), 9090 (gRPC), 9100 (metrics)
```
