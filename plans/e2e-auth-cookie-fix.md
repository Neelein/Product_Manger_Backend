# Plan: Fix E2E auth failures — Secure cookie over HTTP

## Problem
The `session_key` cookie is set with a hardcoded `Secure: true` flag. The backend runs over plain HTTP in development and E2E (`main.go` listens on `:8090`, no TLS; frontend served on `http://localhost:5173` and proxied). Browsers reject `Secure` cookies over non-HTTPS origins, so the session cookie is never stored.

Consequence: login/register appears to succeed (the frontend keeps the member in React state and routes to `/home`), but the first full page reload (`page.goto(...)`) triggers `AuthContext` -> `getCurrentMember()` -> `/api/members/me` with no cookie -> 401 -> redirect to `/login`. This caused 9 Playwright E2E specs to fail/timeout (auth, product, inventory, announcement, calendar, chat).

## Goal
Make the `Secure` flag conditional on the request being HTTPS, so the cookie is stored over HTTP in dev/E2E while still protected over HTTPS in production. No frontend change required.

## Changes
1. New `src/api/session_cookie.go`:
   - `isHTTPS(r)` = `r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")`
   - `sessionCookie(r, session)` — same fields as before but `Secure: isHTTPS(r)`
   - `clearSessionCookie(r)` — logout cookie, `Secure: isHTTPS(r)`
2. `src/api/member_handler.go`: login (line ~102) and logout (line ~125) use the helpers.
3. `src/api/middleware.go`: session-refresh cookie (line ~47) uses the helper.
4. Tests:
   - `src/api/session_cookie_test.go` (unit, no build tag): Secure flag for plain HTTP (false), TLS (true), `X-Forwarded-Proto: https` (true) / `http` (false), clear cookie.
   - `src/test/api/member_handler_test.go` (integration): assert `Secure == false` on plain-HTTP login; add `TestHandler_Login_HTTPSecure`.

## Deliverables
- `plans/e2e-auth-cookie-fix.md` (this file)
- `decisions/2026-08-07-4.md`

## Validation
- `go build ./...`
- `go test ./src/api/... ./src/domain/...`
- Integration tests (`go test -tags=integration ./src/test/api/...`) against Postgres
- Playwright E2E suite: `frontend/scripts/start-backend.sh` + `npx playwright test` — all 9 previously-failing specs should pass.

## Status
- [x] Implement helpers and replace hardcoded cookies
- [x] Add unit + integration tests
- [x] go build + unit tests pass
- [x] Integration tests: `go test -tags=integration -count=1 ./src/test/...` — all packages (api, database, domain) pass together
- [x] Playwright E2E suite — all 9 previously-failing specs now pass

Note: the pre-existing `src/test/database` flakiness (two test packages racing on the shared `productdb_test`) was also fixed so the whole suite is green — see `decisions/2026-08-07-5.md`.