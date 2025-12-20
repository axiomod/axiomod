# Axiomod: Enterprise Go Framework

[![Go Version](https://img.shields.io/github/go-mod/go-version/axiomod/axiomod)](https://golang.org/doc/devel/release.html)
![Build Status](https://img.shields.io/badge/build-passing-brightgreen)
![Coverage](https://img.shields.io/badge/coverage-85%25-green)
[![License](https://img.shields.io/github/license/axiomod/axiomod)](LICENSE)

**Axiomod** is a modular, plugin-based, and enterprise-ready framework for building scalable Go services. Built on Clean Architecture principles, it solves the "blank canvas" problem by providing a robust foundation with built-in observability, authentication, and tooling.

---

## 🚀 Why Axiomod?

- **🧩 Modular Design**: Uses `go.uber.org/fx` for dependency injection, making components loosely coupled and easily testable.
- **🔌 Plugin System**: Extend functionality dynamically without touching the core. Built-in support for DBs, Auth, and more.
- **🛡️ Battle-Tested Defaults**: Comes with production-hardened configurations for HTTP (Fiber), gRPC, Zap Logging, and OpenTelemetry.
- **🔍 Observability First**: Metrics, Logs, and Traces are first-class citizens, pre-configured to work out of the box.
- **🛠️ Developer Tooling**: CLI for rapid scaffolding, migrations, and code validation.

## 📦 Features

| Category | Features |
|----------|----------|
| **Core** | Clean Architecture, Fx Dependency Injection, Config Management (Viper) |
| **API** | Fiber v2 (HTTP), gRPC, Middleware Chains, Validators |
| **Data** | MySQL/PostgreSQL Plugins, Connection Pooling, Transaction Management |
| **Auth** | JWT, OIDC (Keycloak), RBAC (Casbin) |
| **Ops** | Health Probes (Liveness/Readiness), Prometheus Metrics, OpenTelemetry Traces |
| **Async** | Kafka Producers/Consumers, Background Worker Pools |

## 🎯 Use Cases

Axiomod is optimized for:

- **Enterprise Microservices**: Built-in observability (tracing/metrics), health probes, and gRPC support make it ready for Kubernetes meshes.
- **Regulated Systems (FinTech/HealthTech)**: Strict Clean Architecture enforcement ensures implementation details never leak into business rules.
- **Modular Monoliths**: Fx modules allow you to build isolated domains within a single binary and split them easily when scaling.
- **Event-Driven Architectures**: First-class support for Kafka consumers/producers and async background workers.

## ⚡ Quick Start

### Prerequisites

- Go 1.24+
- Docker (optional, for dependencies)

### Installation

```bash
# Clone the repository
git clone https://github.com/axiomod/axiomod.git
cd axiomod

# Install dependencies
make deps

# Build the CLI
make build-cli
```

### Creating a New Project

Use the CLI to scaffold a new service:

```bash
./bin/axiomod init my-awesome-service
cd my-awesome-service
go mod tidy
```

### Running the Server

```bash
# Start with default configuration
go run cmd/axiomod-server/main.go

# Or using the built binary
./bin/axiomod-server
```

## 📚 Documentation

Detailed documentation is available in the [`docs/`](./docs) directory.

- **[Getting Started](./docs/developer-guide.md)**: Setup and workflow.
- **[Architecture](./docs/architecture.md)**: Core concepts and design.
- **[Deployment](./docs/deployment-guide.md)**: Docker/K8s guides.
- **[Plugins](./docs/plugin-development-guide.md)**: How to extend the framework.
- **[API Reference](./docs/api-reference.md)**: HTTP & gRPC contracts.

## 🏗️ Architecture

Axiomod follows **Clean Architecture**:

1. **Entities**: Enterprise business rules.
2. **Use Cases**: Application business rules.
3. **Interface Adapters**: Controllers, Gateways, Presenters.
4. **Frameworks & Drivers**: Web, DB, UI, External Interfaces.

Dependencies point **inwards**, ensuring your business logic remains independent of frameworks and drivers.

## 🗺️ Roadmap

We have ambitious plans to evolve Axiomod into the standard for enterprise Go development. functionality.
Check out our **[Detailed Roadmap](docs/roadmap.md)** to see what's coming next, including:

- 🏢 **Multi-Tenancy Support**
- 🔐 **Advanced Security (Vault, mTLS)**
- 🛠️ **CLI 2.0 (OpenAPI Scaffolding, Monorepos)**
- ⚡ **Event-Driven Patterns (Outbox, DLQ)**

## 🤝 Contributing

We welcome contributions! Please see our [Developer Guide](./docs/developer-guide.md) for details on how to set up your environment and submit PRs.

1. Fork the repo.
2. Create your feature branch (`git checkout -b feature/amazing-feature`).
3. Commit your changes (`git commit -m 'Add amazing feature'`).
4. Push to the branch (`git push origin feature/amazing-feature`).
5. Open a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
