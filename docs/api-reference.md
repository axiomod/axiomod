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

Map domain and framework errors to status codes in the handler:
`framework/errors.ToHTTPCode(err)` converts coded framework errors (e.g.
`CodeNotFound` -> `404`), and domain modules map their own errors — see
`errorResponse` in
[`examples/example/delivery/http/example_handler.go`](../examples/example/delivery/http/example_handler.go)
(not-found -> 404, validation -> 400, otherwise 500).

### Worked Example

The registered example module serves a complete CRUD surface under
`/api/v1/examples` (JWT bearer auth required):

| Method | Path | Description |
|---|---|---|
| POST | `/api/v1/examples/` | Create (201) |
| GET | `/api/v1/examples/` | List with `name`, `valueType`, `tag`, `limit`, `offset` query filters |
| GET | `/api/v1/examples/:id` | Get by ID (404 when missing) |
| PUT | `/api/v1/examples/:id` | Update |
| DELETE | `/api/v1/examples/:id` | Delete |

## 2. gRPC API

gRPC is used for high-performance service-to-service communication.

### Contracts (Protobuf)

By convention, each domain owns its contracts in
`<domain>/delivery/grpc/*.proto`. The reference contract is
[`examples/example/delivery/grpc/example.proto`](../examples/example/delivery/grpc/example.proto)
(Create/Get/Update/Delete/List). Regenerate with `protoc` or `buf` using
`protoc-gen-go` and `protoc-gen-go-grpc`:

```proto
syntax = "proto3";
package example.v1;

service ExampleService {
  rpc CreateExample(CreateExampleRequest) returns (CreateExampleResponse);
  rpc GetExample(GetExampleRequest) returns (GetExampleResponse);
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

The framework does not yet generate OpenAPI specifications (planned — see
[the CLI roadmap](roadmap/cli-enhancement.md)). If you author a spec by
hand, validate it with `axiomod validator check-api-spec --spec <file>`.
For generation today, `swaggo/swag` works with Fiber handlers.

### gRPC Reflection

gRPC reflection is enabled by default in development environments, allowing you to use tools like `grpcurl` or Postman to explore the API.

## 4. Design Guidelines

- **Versioning**: Always version your APIs (e.g., `/api/v1/...`).
- **Idempotency**: Ensure that `POST`, `PUT`, and `DELETE` requests are idempotent where possible.
- **Payloads**: Use JSON for HTTP and Protobuf for gRPC. Avoid returning internal implementation details in API models.
