# Clean Architecture Migration

## Goal

Separate core models, repository ports, application use cases, HTTP adapters, and PostgreSQL/session/storage adapters while preserving existing API contracts and behavior.

## Target boundaries

- `src/domain/model`: core entities and value data only.
- `src/domain/repository`: repository and infrastructure ports.
- `src/usecase`: validation, authorization, defaults, and application workflows.
- `src/adapter/http`: request/response translation, routing, and middleware.
- `src/adapter/postgres`: PostgreSQL persistence implementations.
- `src/adapter/session` and `src/adapter/storage`: non-database infrastructure adapters.
- `src/infrastructure`: database pool and migration setup.

## Migration strategy

1. Introduce model and repository packages while preserving behavior.
2. Add use cases for member, product/variant, inventory, and remaining domains.
3. Make HTTP handlers depend on use cases instead of concrete persistence.
4. Move PostgreSQL, session, storage, and migration responsibilities to adapters/infrastructure.
5. Keep existing routes, JSON contracts, migrations, and frontend E2E behavior unchanged.

## Additional corrections

- Use request context instead of `context.Background()` in HTTP flows.
- Protect product update/delete routes with authentication.
- Keep the runtime port at `8090` and align documentation.
- Preserve database constraints and stored functions where they provide atomicity or integrity; move application decisions to use cases.

## Production cutover

1. Remove the legacy `src/api`, `src/database`, and root `src/domain` production packages after migrating all tests to bounded packages.
2. Register routes and HTTP translation from `src/adapter/http`, with handlers depending on application services rather than PostgreSQL concrete types.
3. Put PostgreSQL repository implementations in `src/adapter/postgres`, implementing `domain/repository` ports and using `domain/model` entities.
4. Construct repositories, usecases, session storage, and HTTP adapters in `main.go`; no repository is passed directly to a legacy API package.
5. Preserve route paths, status codes, cookie behavior, and JSON response contracts while moving application validation and authorization behind usecase facades.

## Verification

- Add use case unit tests and dependency architecture tests.
- Run API/database tests against the HTTP, PostgreSQL, session, model, and repository packages directly.
- Run `gofmt`, `go test ./...`, and the available database integration tests. E2E is intentionally deferred until backend verification is complete.
