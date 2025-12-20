# API Reference & Contracts

Axiomod supports both HTTP and gRPC protocols for inter-service communication and edge interaction.

## 1. HTTP API (Fiber)

The framework uses [Fiber v2](https://gofiber.io/) for its HTTP server.

### defining Handlers

Handlers are typically located in the `delivery/http` directory of your service.

```go
func (h *Handler) GetUser(c *fiber.Ctx) error {
    id := c.Params("id")
    user, err := h.useCase.GetUser(c.Context(), id)
    if err != nil {
        return err // The error handler middleware converts this to a response
    }
    return c.JSON(user)
}
```

### Error Handling

The framework includes a centralized error handler that maps internal framework errors (from `github.com/axiomod/axiomod/framework/errors`) to appropriate HTTP status codes (e.g., `ErrNotFound` -> `404`).

## 2. gRPC API

gRPC is used for high-performance service-to-service communication.

### Contracts (Protobuf)

Define your API contracts in the `api/` directory using `.proto` files.

```proto
syntax = "proto3";
package axiomod.v1;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
}
```

### Implementing the Server

Implement the generated interface in `delivery/grpc`:

```go
func (s *Server) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.GetUserResponse, error) {
    user, err := s.useCase.GetUser(ctx, req.Id)
    if err != nil {
        return nil, status.Error(codes.NotFound, err.Error())
    }
    return &v1.GetUserResponse{User: user}, nil
}
```

## 3. API Documentation

### OpenAPI / Swagger

For HTTP APIs, it is recommended to use `swaggo/swag` or similar tools to generate OpenAPI specifications from comments in your code.

### gRPC Reflection

gRPC reflection is enabled by default in development environments, allowing you to use tools like `grpcurl` or Postman to explore the API.

## 4. Design Guidelines

- **Versioning**: Always version your APIs (e.g., `/api/v1/...`).
- **Idempotency**: Ensure that `POST`, `PUT`, and `DELETE` requests are idempotent where possible.
- **Payloads**: Use JSON for HTTP and Protobuf for gRPC. Avoid returning internal implementation details in API models.
