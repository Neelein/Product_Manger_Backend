# Production Runtime Cutover

## Goal

Make the production dependency path `main -> adapter/http -> usecase -> repository -> adapter/postgres` while preserving routes and response contracts.

## Steps

1. Add bounded-context usecase service interfaces and repository-backed implementations.
2. Change HTTP handlers and middleware constructors to accept usecase services and infrastructure ports only.
3. Make the HTTP router compose handlers from services, and make `main.go` the only runtime composition root.
4. Add import-boundary architecture tests and retain legacy packages only for compatibility tests.
5. Run formatting, unit tests, and requested integration test commands.
