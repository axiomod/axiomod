# Deployment Guide

## Introduction

This guide provides instructions for deploying the Enterprise Axiomod in various environments. The framework is designed to be deployed in containers and supports various deployment options.

## Prerequisites

Before deploying the framework, ensure you have the following:

- Go 1.24 or higher
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

Create a `config.yaml` file with your configuration:

```yaml
app:
  name: axiomod
  environment: production
  version: 1.0.0
  debug: false

http:
  host: 0.0.0.0
  port: 8080
  readTimeout: 30
  writeTimeout: 30
  shutdownTimeout: 30

grpc:
  host: 0.0.0.0
  port: 9090
  shutdownTimeout: 30

database:
  driver: mysql
  host: mysql
  port: 3306
  username: root
  password: password
  database: axiomod
  maxOpenConns: 10
  maxIdleConns: 5
  connMaxLifetime: 300


kafka:
  brokers:
    - kafka:9092
  groupId: axiomod

auth:
  provider: jwt
  jwtSecret: your-secret-key
  jwtDuration: 3600

observability:
  logLevel: info
  logFormat: json
  metricsEnabled: true
  metricsPort: 9100
  tracingEnabled: true
  tracingServiceName: axiomod
  tracingExporterType: jaeger
  tracingExporterURL: http://jaeger:14268/api/traces

plugins:
  enabled:
    - mysql
    - jwt
  config:
    mysql:
      maxRetries: 3
    jwt:
      algorithm: HS256
```

### Environment Variables

You can override configuration values using environment variables:

```bash
# App configuration
export APP_NAME=axiomod
export APP_ENV=production
export APP_VERSION=1.0.0
export APP_DEBUG=false

# HTTP configuration
export HTTP_HOST=0.0.0.0
export HTTP_PORT=8080

# Database configuration
export DB_DRIVER=mysql
export DB_HOST=mysql
export DB_PORT=3306
export DB_USERNAME=root
export DB_PASSWORD=password
export DB_DATABASE=axiomod

# And so on for other configuration values...
```

## Deployment Options

### Docker Compose

For local development or simple deployments, you can use Docker Compose:

```yaml
# docker-compose.yml
version: '3'

services:
  app:
    image: axiomod-framework:latest
    ports:
      - "8080:8080"
      - "9090:9090"
      - "9100:9100"
    environment:
      - APP_ENV=production
      - DB_HOST=mysql
      - REDIS_HOST=redis
      - KAFKA_BROKERS=kafka:9092
    depends_on:
      - mysql
      - kafka

  mysql:
    image: mysql:8.0
    environment:
      - MYSQL_ROOT_PASSWORD=password
      - MYSQL_DATABASE=axiomod
    volumes:
      - mysql-data:/var/lib/mysql


  kafka:
    image: confluentinc/cp-kafka:7.0.0
    environment:
      - KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://kafka:9092
      - KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1
    volumes:
      - kafka-data:/var/lib/kafka/data

volumes:
  mysql-data:
  kafka-data:
```

To start the services:

```bash
docker-compose up -d
```

### Kubernetes

For production deployments, you can use Kubernetes:

```yaml
# kubernetes/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: axiomod
  labels:
    app: axiomod
spec:
  replicas: 3
  selector:
    matchLabels:
      app: axiomod
  template:
    metadata:
      labels:
        app: axiomod
    spec:
      containers:
      - name: axiomod
        image: axiomod:latest
        ports:
        - containerPort: 8080
        - containerPort: 9090
        - containerPort: 9100
        env:
        - name: APP_ENV
          value: production
        - name: DB_HOST
          value: mysql
        - name: KAFKA_BROKERS
          value: kafka:9092
        resources:
          limits:
            cpu: "1"
            memory: "512Mi"
          requests:
            cpu: "0.5"
            memory: "256Mi"
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 15
          periodSeconds: 20
```

```yaml
# kubernetes/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: axiomod
spec:
  selector:
    app: axiomod
  ports:
  - name: http
    port: 8080
    targetPort: 8080
  - name: grpc
    port: 9090
    targetPort: 9090
  - name: metrics
    port: 9100
    targetPort: 9100
  type: ClusterIP
```

To deploy to Kubernetes:

```bash
kubectl apply -f kubernetes/
```

## Scaling

The framework is designed to be horizontally scalable. You can scale the application by:

1. Increasing the number of replicas in Kubernetes
2. Using a load balancer to distribute traffic
3. Ensuring all stateful components (databases, caches, etc.) are properly scaled

## Monitoring and Observability

The framework provides built-in support for monitoring and observability:

- **Metrics**: Exposed on port 9100 in Prometheus format
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
      - targets: ['axiomod:9100']
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
