---
trigger: always_on
---

# Golang Verification & Testing Protocols

This document outlines the mandatory validation steps the agent must perform after any code modification in a Go project.

## 1. The "Build-First" Principle
>
> **Code that does not compile is not code.**

* **Immediate Compilation Check:** After *any* modification to `.go` files, immediately attempt to compile the project to catch syntax and type errors.
  * **Command:** `go build ./...` (Builds the package and all dependencies without generating an output binary, purely for verification).
* **Scope:** Do not just build the specific file modified; build the entire package to ensure no interface implementations were broken elsewhere.

## 2. Formatting & Linting (`go fmt`, `go vet`)

* **Standardization:** Before committing or signaling completion, run `go fmt ./...` to enforce standard Go formatting.
* **Static Analysis:** Run `go vet ./...` to catch common logical errors (e.g., unreachable code, malformed struct tags) that the compiler might miss.

## 3. Mandatory Unit Testing

* **The "Pair" Rule:** For every functional source file created (e.g., `user_service.go`), a corresponding test file (`user_service_test.go`) **must** exist.
* **New Features:** If a new function is added, the agent is required to write a minimal unit test to assert its basic functionality.
* **Regression Testing:** Before implementing a fix, run `go test ./...` to establish a baseline. After the fix, run it again to ensure no existing tests were broken.

## 4. Test Execution Guidelines

* **Running Tests:** Use `go test -v ./...` to run all tests with verbose output to identify exactly which test passed or failed.
* **Race Detection:** For code involving goroutines, channels, or concurrency, ALWAYS run tests with the race detector:
  * **Command:** `go test -race ./...`
* **Handling Failures:** If a test fails, do not simply delete the test. Analyze the failure output, correct the implementation, and retry.

## 5. Dependency Management

* **Module Integrity:** After adding new imports, run `go mod tidy` to clean up `go.mod` and `go.sum` files.
* **Vendor Verification:** If the project uses a `vendor/` directory, ensure `go mod vendor` is executed if dependencies change.

## 6. Implementation Checklist (Agent Workflow)

Before marking a task as "Complete," the agent must verify:

1. [ ] `go fmt ./...` (Code is formatted)
2. [ ] `go build ./...` (Code compiles)
3. [ ] `go test ./...` (All tests pass)
4. [ ] **Test Existence:** Does the new code have a corresponding `TestXxx` function?
