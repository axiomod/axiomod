# gRPC Delivery

## Service Structure

```go
type ExampleGRPCService struct {
    createUseCase *usecase.CreateExampleUseCase
    getUseCase    *usecase.GetExampleUseCase
    logger        *observability.Logger
    UnimplementedExampleServiceServer // Embed for forward compatibility
}

func NewExampleGRPCService(
    createUseCase *usecase.CreateExampleUseCase,
    getUseCase    *usecase.GetExampleUseCase,
    logger        *observability.Logger,
) *ExampleGRPCService { ... }
```

## Method Pattern

```go
func (s *ExampleGRPCService) CreateExample(ctx context.Context,
    req *CreateExampleRequest) (*CreateExampleResponse, error) {
    // 1. Map request to use case input
    input := usecase.CreateExampleInput{
        Name:        req.Name,
        Description: req.Description,
    }

    // 2. Execute use case
    output, err := s.createUseCase.Execute(ctx, input)
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }

    // 3. Map to response
    return &CreateExampleResponse{Id: output.ID}, nil
}
```

## Interceptor Chain (server-level)

Applied in order: `ctxtags -> zap logging -> validation -> recovery -> metrics -> tracing -> timeout`

Optional: auth interceptor, RBAC interceptor.

## gRPC Server Features

- Health service registered automatically
- Reflection registered automatically
- TLS support (configurable)
- Keepalive parameters
- Graceful shutdown via `GracefulStop()`

## Interceptors

| Interceptor | Purpose |
|---|---|
| `MetricsInterceptor` | Prometheus counters + histograms per service/method |
| `TracingInterceptor` | OTel spans with rpc attributes |
| `RBACInterceptor` | Casbin policy enforcement using `info.FullMethod` |
| `timeoutInterceptor` | Context timeout wrapper |
| `recoveryHandler` | Panic recovery returning `codes.Internal` |

## Registration

gRPC services registered via `fx.Invoke`:

```go
func registerGRPCServices(grpcServer *grpc.Server, service *grpc.ExampleGRPCService) {
    RegisterExampleServiceServer(grpcServer, service)
}
```

## Rules

1. Same as HTTP: **no business logic** in gRPC methods -- delegate to use cases
2. Always embed `Unimplemented*Server` for forward compatibility
3. Map gRPC requests to use case inputs, use case outputs to gRPC responses
4. Use `status.Error(codes.X, msg)` for gRPC error responses
5. Use `framework/errors.ToGRPCCode(err)` for automatic mapping
