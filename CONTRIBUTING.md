# Contributing to Orkai Observability

Thank you for your interest in contributing to Orkai Observability! We welcome contributions of all forms, including bug reports, feature requests, documentation improvements, and pull requests.

## How to Contribute

### 1. Reporting Bugs & Feature Requests
* Please open a new Issue describing the bug or the feature request in detail.
* Provide a clear description, reproduction steps, and the expected behavior.

### 2. Pull Request Guidelines
* **Create a Branch:** Create a branch on your fork named after the feature/bug (e.g., `feat/new-metric`, `fix/stack-race`).
* **Follow Code Guidelines:** 
  * All public types, functions, and interfaces **must have comprehensive Go Doc comments** containing clear intents and a practical `Usage example:`.
  * Keep functions focused, clean, and highly readable.
  * Adhere to the Single Responsibility Principle (SRP).
* **Write Tests:** Every bug fix and new feature must be backed by fast, isolated, and reliable unit tests in the `/test` directory.
* **Run Formatter:** Ensure your code is formatted with the official `gofmt` before committing:
  ```bash
  go fmt ./...
  ```
* **Verify Tests:** Ensure all tests pass successfully before opening a Pull Request:
  ```bash
  go test ./... -v
  ```

## Development Environment
To run the project locally, ensure you have Go 1.20+ installed. Clones and forks can be set up in any local workspace.
