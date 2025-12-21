# ADR-004: Integrated RBAC via Casbin

## Status

Accepted

## Context

Managing access control across various HTTP and gRPC services was previously fragmented and required manual logic. There was no standardized way to enforce fine-grained permissions within the framework.

## Decision

We have adopted Casbin as the primary library for Role-Based Access Control (RBAC).

- Implemented `RBACService` to wrap Casbin's enforcer.
- Created standard middleware for Fiber and interceptors for gRPC to automate enforcement.
- Support for external policy files (CSV/YAML) and future database adapters.
- Integration with the global configuration system for model and policy paths.

## Consequences

- **Consistency**: Standardized authorization logic across the entire framework.
- **Flexibility**: Casbin's support for various model types (RBAC, ABAC, ACL) allows for future expansion.
- **Performance**: Policy evaluation is highly optimized with Casbin; however, very large policy sets may require careful management.
- **Dependency**: Introduced a new dependency on `github.com/casbin/casbin/v2`.
