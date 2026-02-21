---
description: "Scaffold a complete Clean Architecture domain module with all 8 packages. Use when the user says 'create module', 'new domain', 'scaffold module', 'generate module <name>'."
---

# Generate Domain Module

## Inputs
- Module name (required): lowercase, alphanumeric, no hyphens (Go package naming)

## Steps

1. **Validate the module name**:
   - Must be lowercase, alphanumeric (Go package naming convention)
   - Check that `examples/<name>/` does not already exist

2. **Run the CLI generator**:
   ```bash
   axiomod generate module --name=<name>
   ```

3. **Verify generated structure** (8 packages + module.go):
   ```bash
   ls -R examples/<name>/
   ```
   Expected: entity/, repository/, usecase/, service/, delivery/http/, delivery/grpc/, infrastructure/persistence/, infrastructure/cache/, infrastructure/messaging/, module.go

4. **Post-generation tasks** (inform the user):
   - Implement entity fields and validation in `entity/<name>.go`
   - Define repository methods in `repository/<name>_repository.go`
   - Add business logic to use cases in `usecase/`
   - Wire into application by adding `<name>.Module` to `cmd/axiomod-server/fx_options.go`
   - Add architecture rules for the new module to `architecture-rules.json` if needed
   - Run `axiomod validator architecture` to verify no import violations

5. **Optional enhancements**:
   - Generate additional use cases: `axiomod generate service --name=<service> --module=<name>`
   - Generate additional handlers: `axiomod generate handler --name=<handler> --module=<name>`
   - Add tests following `examples/example/example_test.go` pattern
