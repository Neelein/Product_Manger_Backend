# Plan — Ordered CI/CD pipeline: dev/PR → tests → merge → tests → deploy (2026-08-08)

## Goal
A single backend pipeline where merge and deploy are gated on ordered test stages, and the E2E stage tests the branch-matched frontend (dev ↔ dev, main ↔ main).

## Flow
1. Push to `dev` or PR (dev → main): `unit` → `integration` → `e2e`, in strict order — must all pass before merge.
2. Merge to `main`: run `unit` → `integration` → `e2e` again; only then `deploy` (GHCR image + server).

## Changes
| File | Action | Description |
|------|--------|-------------|
| `backend/.github/workflows/ci.yml` | Rewrite | `unit` (go test) → `integration` (needs unit, `go test -tags=integration`) → `e2e` (needs integration, frontend checkout at `${{ github.ref_name }}`) → `deploy` (needs e2e, `if: github.ref == 'refs/heads/main'`) |
| `backend/.github/workflows/e2e.yml` | Delete | Folded into `ci.yml` |

## Triggers
```yaml
on:
  push: { branches: [main, dev] }
  pull_request: { branches: [main] }
```

## Merge enforcement
Add `unit`/`integration`/`e2e` as required status checks under branch protection in GitHub repo settings.

## Verification
- YAML syntax validated.
- Push to `dev` confirms the dev chain; a `main` merge confirms full chain + deploy.