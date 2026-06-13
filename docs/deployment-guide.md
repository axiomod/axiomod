# Deployment Guide

## Introduction

This guide provides instructions for deploying the Enterprise Axiomod in various environments. The framework is designed to be deployed in containers and supports various deployment options.

## Prerequisites

Before deploying the framework, ensure you have the following:

- Go 1.25 or higher
- Docker and Docker Compose (for containerized deployment)
- Kubernetes (for orchestrated deployment)
- Access to required infrastructure services (databases, message brokers, etc.)

## Building the Application

### 1. Build the binary

```bash
# Navigate to the project directory
cd axiomod

# Build the binary
go build -o bin/axiomod-server ./cmd/axiomod-server
```

### 2. Build the Docker image

```bash
# Build the Docker image
docker build -t axiomod:latest .
```

## Configuration

The framework uses a hierarchical configuration system that can be configured through:

- YAML/JSON files
- Environment variables
- Command-line flags

### Configuration File

The canonical configuration file is `configs/service_default.yaml`; the
loader searches `configs/`, `config/`, and the working directory, and the
Docker image passes `-config /app/configs/service_default.yaml` explicitly.
A production configuration looks like:

```yaml
app:
  name: "axiomod-service"
  environment: "production"
  version: "0.2.0"
  debug: false

http:
  host: "0.0.0.0"
  port: 8080
  readTimeout: 10
  writeTimeout: 10

grpc:
  host: "0.0.0.0"
  port: 9090

database:
  driver: "mysql" # or postgres
  host: "mysql"
  port: 3306
  user: "axiomod"
  password: "" # inject via APP_DATABASE_PASSWORD
  name: "axiomod"
  sslMode: "require"
  orm: "ent" # ent (default) | sql
  maxOpenConns: 25
  maxIdleConns: 5
  connMaxLifetime: 15 # minutes
  slowQueryThreshold: 200 # milliseconds

auth:
  jwt:
    secretKey: "" # inject via APP_AUTH_JWT_SECRETKEY (min 32 bytes)
    tokenDuration: 60 # minutes

observability:
  logLevel: "info"
  logFormat: "json"
  metricsEnabled: true # served at /metrics on the HTTP port
  tracingEnabled: true
  tracingExporterType: "otlp" # jaeger | otlp | stdout
  tracingUrl: "jaeger:4317"
  tracingSamplerRatio: 0.1

plugins:
  enabled: # map of plugin name -> bool
    mysql: true
    jwt: true
    kafka: true
  settings: # per-plugin settings blocks
    kafka:
      brokers: ["kafka:9092"]
      clientId: "axiomod"
```

### Environment Variables

Every Viper key can be overridden with an environment variable: prefix
`APP_`, replace key dots with underscores (the camelCase part stays joined).
Overrides apply to keys present in the configuration file.

```bash
# Viper key                  -> environment variable
# app.environment            -> APP_APP_ENVIRONMENT
# http.port                  -> APP_HTTP_PORT
# database.host              -> APP_DATABASE_HOST
# database.password          -> APP_DATABASE_PASSWORD
# auth.jwt.secretKey         -> APP_AUTH_JWT_SECRETKEY
# observability.logLevel     -> APP_OBSERVABILITY_LOGLEVEL
# plugins.enabled.postgres   -> APP_PLUGINS_ENABLED_POSTGRES

export APP_APP_ENVIRONMENT=production
export APP_DATABASE_HOST=mysql
export APP_DATABASE_PASSWORD=change-me
export APP_AUTH_JWT_SECRETKEY=$(openssl rand -hex 32)
```

## Deployment Options

### Docker Compose

A complete reference stack (app, PostgreSQL, Redis, Kafka/Zookeeper, Jaeger,
Prometheus) ships at
[`docker-compose.reference.yaml`](../docker-compose.reference.yaml):

```bash
docker compose -f docker-compose.reference.yaml up -d
```

The app service demonstrates the real `APP_*` environment overrides
(`APP_DATABASE_HOST=postgres`, `APP_PLUGINS_ENABLED_POSTGRES=true`, …).
Ports: 8080 (HTTP + `/metrics`), 9090 (gRPC); Prometheus UI is mapped to
host port 9091, Jaeger UI to 16686.

### Kubernetes

Production-ready manifests ship in
[`deploy/kubernetes/`](../deploy/kubernetes/): `deployment.yaml` (3 replicas,
non-root security context, `/ready` and `/live` probes, secret-injected
credentials), `service.yaml` (ClusterIP for http/grpc), and `configmap.yaml`
(the mounted `service_default.yaml`).

```bash
# Create the secrets the deployment references
kubectl create secret generic axiomod-secrets \
  --from-literal=DB_PASSWORD=your-db-password \
  --from-literal=JWT_SECRET=$(openssl rand -hex 32)

kubectl apply -f deploy/kubernetes/
```

## Scaling

The framework is designed to be horizontally scalable. You can scale the application by:

1. Increasing the number of replicas in Kubernetes
2. Using a load balancer to distribute traffic
3. Ensuring all stateful components (databases, caches, etc.) are properly scaled

## Monitoring and Observability

The framework provides built-in support for monitoring and observability:

- **Metrics**: Exposed at `/metrics` on the HTTP port (8080) in Prometheus format
- **Logging**: Structured JSON logs
- **Tracing**: Distributed tracing with OpenTelemetry

### Prometheus and Grafana

You can use Prometheus to scrape metrics and Grafana to visualize them:

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'axiomod'
    scrape_interval: 15s
    static_configs:
      - targets: ["axiomod:8080"]
```

### ELK Stack

You can use the ELK stack (Elasticsearch, Logstash, Kibana) to collect and analyze logs:

```yaml
# logstash.conf
input {
  tcp {
    port => 5000
    codec => json
  }
}

filter {
  # Add filters as needed
}

output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "axiomod-%{+YYYY.MM.dd}"
  }
}
```

### Jaeger

You can use Jaeger to collect and visualize traces:

```yaml
# docker-compose.yml (excerpt)
services:
  jaeger:
    image: jaegertracing/all-in-one:1.30
    ports:
      - "16686:16686"
      - "14268:14268"
```

## Security Considerations

### 1. Secrets Management

Use a secrets management solution like Kubernetes Secrets, HashiCorp Vault, or AWS Secrets Manager to manage sensitive information:

```bash
# Create a Kubernetes secret
kubectl create secret generic axiomod-secrets \
  --from-literal=DB_PASSWORD=your-db-password \
  --from-literal=JWT_SECRET=your-jwt-secret
```

### 2. Network Security

Use network policies to restrict communication between services:

```yaml
# kubernetes/network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: axiomod-network-policy
spec:
  podSelector:
    matchLabels:
      app: axiomod
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: frontend
    ports:
    - protocol: TCP
      port: 8080
    - protocol: TCP
      port: 9090
```

### 3. TLS

Use TLS to encrypt communication:

```yaml
# kubernetes/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: axiomod-ingress
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - api.example.com
    secretName: api-tls
  rules:
  - host: api.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: axiomod
            port:
              number: 8080
```

## Conclusion

This deployment guide provides a starting point for deploying the Enterprise Axiomod. Depending on your specific requirements, you may need to adjust the configuration and deployment options.
