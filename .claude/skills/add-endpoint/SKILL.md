---
description: "Add a new HTTP + gRPC endpoint to an existing domain module. Use when the user says 'add endpoint', 'new route', 'add API', 'new handler method'."
---

# Add Endpoint to Existing Module

## Inputs
- Module name (required): existing module in examples/
- Action name (required): e.g., "update", "delete", "list"
- HTTP method and path (optional, inferred from action)
- Whether gRPC counterpart is needed (default: yes)

## Steps

1. **Add use case** in `examples/<module>/usecase/<action>_<module>.go`:
   ```go
   type <Action><Module>Input struct { ... }
   type <Action><Module>Output struct { ... }
   type <Action><Module>UseCase struct { repo repository.<Module>Repository }
   func New<Action><Module>UseCase(repo repository.<Module>Repository) *<Action><Module>UseCase { ... }
   func (uc *<Action><Module>UseCase) Execute(ctx context.Context, input <Action><Module>Input) (*<Action><Module>Output, error) { ... }
   ```

2. **Add repository method** if needed in `examples/<module>/repository/<module>_repository.go`

3. **Update infrastructure** implementations to satisfy new repository methods

4. **Add handler method** in `examples/<module>/delivery/http/<module>_handler.go`:
   - Add dependency field for the new use case
   - Update constructor to accept new use case
   - Add route in `RegisterRoutes()`
   - Implement handler method

5. **Add gRPC method** (if needed) in `examples/<module>/delivery/grpc/<module>_grpc_service.go`:
   - Add method to service struct
   - Implement the RPC method
   - Update proto file if applicable

6. **Update module.go wiring**:
   ```go
   fx.Provide(usecase.New<Action><Module>UseCase),
   ```

7. **Add tests**:
   - Unit test for the use case
   - Handler test with httptest if applicable

8. **Verify**:
   ```bash
   go build ./examples/<module>/...
   go test ./examples/<module>/...
   axiomod validator architecture
   ```
