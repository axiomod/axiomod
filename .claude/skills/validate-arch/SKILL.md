---
description: "Run architecture validation against import rules. Use when the user says 'validate architecture', 'check imports', 'verify layers', 'run arch check'."
---

# Validate Architecture

## Steps

1. **Run the CLI validator**:
   ```bash
   axiomod validator architecture
   ```
   This checks all Go imports against `architecture-rules.json`.

2. **If violations are found**, analyze each:
   - Read the violating file
   - Identify the forbidden import
   - Suggest the correct dependency direction
   - Propose refactoring (e.g., extract interface, move code to correct layer)

3. **Additional validation with custom config**:
   ```bash
   axiomod validator architecture --config=architecture-rules.json
   ```

4. **Manual verification for edge cases**:
   - Check that no domain module imports another domain module
   - Verify pattern rules (internal/*/entity may import framework/*)
   - Confirm exceptions are valid (_test.go, mock_, testdata)

## Architecture Rules Summary

- entity: no domain imports
- repository -> entity
- usecase -> entity, repository, service
- service -> entity, repository
- delivery/http -> usecase, entity, middleware
- delivery/grpc -> usecase, entity
- infrastructure/* -> entity, repository
- platform/* -> framework/*
- plugins/* -> platform/*, framework/*
- No cross-domain imports
