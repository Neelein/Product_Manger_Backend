# Ordered CI/CD pipeline (2026-08-08)

## What
Consolidated the backend workflows into a single `ci.yml` with ordered stages: `unit` → `integration` → `e2e` → `deploy`. Removed the separate `e2e.yml`.

## Why
The user wants a dev/PR flow where 單元 → 整合 → e2e tests run strictly in order and must pass before merging, followed by the same chain after a `main` merge with deploy only after the full chain passes.

## Changes
- `backend/.github/workflows/ci.yml`:
  - `unit`: `go test -count=1 ./src/...`.
  - `integration` (needs unit): postgres service `productdb_test`, `go test -tags=integration -count=1 -p=1 -v ./src/test/...`.
  - `e2e` (needs integration): full-stack Playwright; frontend checked out at `${{ github.ref_name }}` (dev ↔ dev, main ↔ main); backend under test is the repo's own checkout.
  - `deploy` (needs e2e): `if: github.ref == 'refs/heads/main'` — GHCR image + server deploy only after a green chain on a merge.
- Deleted `backend/.github/workflows/e2e.yml`.

## Notes
- On a PR, `github.ref_name` is the head branch (dev); on a `main` push it is `main`, so the E2E partner stays branch-matched.
- Branch protection in repo settings must require `unit`/`integration`/`e2e` checks to actually block merges.

## Verification
- YAML syntax validated for `ci.yml`.
- Push to `dev` and watch the ordered dev pipeline; a `main` merge runs the full chain then deploys.