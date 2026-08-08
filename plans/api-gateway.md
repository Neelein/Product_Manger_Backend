# Plan — API gateway: shared Bearer secret enforced on /api/* (2026-08-08)

## Goal
Only callers that present `Authorization: Bearer <SECRET>` can hit the backend `/api/*` routes, so external scripts (direct calls to `neeleindev.com:8090` without the secret) are rejected. The frontend is the legitimate caller: in production the secret is injected by a Vercel serverless/edge proxy, NEVER embedded in the browser bundle.

## Decisions (confirmed with user)
- Scope: `/api/*` only (excluding `/api/health`). `/media/*` and `/` stay public.
- Frontend approach = Option B: secret lives on Vercel + backend, not in the bundle.
- Backend behavior when `API_GATEWAY_SECRET` unset: gate OFF + startup warning (local/dev-friendly).
- Env naming: `API_GATEWAY_SECRET` (backend, Vercel, and local Vite server proxy); no `VITE_*` secret is used.
- 401 body: `{"error":"unauthorized"}` (existing shape).

## Changes
| File | Action | Description |
|------|--------|-------------|
| `src/api/gate_middleware.go` | New | `GatewayMiddleware(secret)` — pass through when secret empty; exempt non-`/api` and `/api/health`; constant-time compare of `Authorization: Bearer <token>`; 401 otherwise |
| `main.go` | Modify | Read `API_GATEWAY_SECRET`, wrap mux with gateway, log warning when unset |
| `dotenv.env` | Modify | Document `API_GATEWAY_SECRET` |
| `src/test/api/gate_middleware_test.go` | New | Unit table-driven: valid/invalid/missing header, health + non-api exemptions, empty secret noop |
| `.github/workflows/cicd.yml` | Modify | Deploy `docker run` adds `-e API_GATEWAY_SECRET=${{ secrets.API_GATEWAY_SECRET }}` |
| `plans/api-gateway.md`, `decisions/2026-08-08-api-gateway.md` | New | Docs |

## Verification
- `gofmt -w` on changed Go files; `go test ./...`; `go vet ./...`
- `go test -tags=integration -count=1 -p=1 ./src/test/...`
- Full Playwright E2E run with the gate actually enabled (E2E_API_SECRET wired through both webServers).

## Limitations
- Ready Server remains HTTP on port 8090; Bearer credentials can be intercepted in transit. A future HTTPS deployment must rotate the shared secret.
- `/media/*` remains an HTTP rewrite to the existing backend and is not routed through the gateway.
