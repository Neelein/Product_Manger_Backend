# Registration code — backend implementation

Boundary: restricted signup via one-time registration codes issued by an admin.

## Current status
- Backend fully implemented and tested (build + integration tests green).
- Frontend implemented (register code field, admin registration-codes page + route + nav).
- E2E: 12/12 pass (incl. 3 new registration-code specs).
- Pending: commit + push to `dev` (requires approval).

## Decisions
- New `members.role` column (`'member'` default, `'admin'`). Admin seeded in 014 has `role='admin'` (promoted in 015).
- Registration now requires a one-time-use code validated atomically in `register_member_with_code` (`FOR UPDATE` consumes the code in-txn).
- New admin-only endpoints under `/api/registration-codes` (POST create, GET list, DELETE {id}), guarded by `RequireRole("admin", auth)`.
- Each code is single-use: consumed at registration and cannot be reused.
- Code auto-generates (8 chars) when the admin leaves it blank; admin may supply a custom string.

## Backend files
- `db/migrations/015_create_registration_codes.up.sql` — role column, seed promote, table + 4 plpgsql/sql functions.
- `src/domain/{member.go, repository.go, errors.go}` — types, `RegistrationCodeRepository` interface, sentinel errors.
- `src/database/registration_code_repo.go` — repo.
- `src/database/member_repo.go` — 7-column scan for member lookups.
- `src/api/registration_code_handler.go`, `member_handler.go`, `middleware.go`, `router.go`, `main.go` — wiring + `RequireRole`.

## Bugs found & fixed during testing
- `CREATE OR REPLACE` cannot change a function's return type (get_member_by_email / get_member_by_id 6→7 cols) → DROP + CREATE.
- PL/pgSQL `RETURNING id, ...` and `WHERE id =` ambiguous with `RETURNS TABLE` OUT columns → table-qualify.
- Literal `'member'` is `text`, not `varchar` → cast.
- `DELETE ... RETURNING TRUE INTO v` becomes NULL on zero rows → IS NULL guard.
- `list_registration_codes` returned NULL UUID/email that pgx cannot scan into `string` → COALESCE to ''.

## Tests
- `src/test/api/member_handler_test.go`: register with valid/missing/invalid/used codes, duplicate email, login, /me (now asserts `role`), idle-expiry, logout.
- `src/test/api/registration_code_handler_test.go`: admin create/list/delete, auto-gen, 403 for members, 401 unauthenticated, delete-not-found 404.

## Next
1. Frontend: register form code field (`LoginPage`), `registrationCode` API client, role in auth context, admin `RegistrationCodesPage`, routes + `Layout` nav.
2. E2E helpers seed a code; new specs (admin create/list/delete; invalid/used code).
3. Final full E2E run; commit + push to `dev` (with approval).