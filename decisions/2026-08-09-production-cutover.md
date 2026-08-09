# Production Cutover Decision

The runtime path is being switched from the legacy API/database packages to the Clean Architecture adapters. HTTP handlers and DTOs live under `adapter/http`, PostgreSQL implementations live under `adapter/postgres`, and `main.go` composes repositories, usecases, session storage, and HTTP routes. Legacy packages remain only as compatibility surfaces for existing tests so API behavior and test contracts remain stable.
