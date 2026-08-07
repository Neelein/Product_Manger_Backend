# Plan — Backend E2E workflow (2026-08-07)

## Overview

Add a GitHub Actions E2E workflow to the backend repo. It reuses the Playwright full-stack E2E suite whose single source of truth lives in the frontend repo: it checks out the frontend (`dev`), serves the frontend, builds + runs this backend checkout against a dedicated `productdb_e2e` Postgres, and executes the same Playwright suite. This gives E2E coverage when the backend itself changes.

## Why

The frontend repo already runs E2E when its code changes. The backend must also validate cross-repo compatibility on every push to `main`/`dev`, without duplicating the Playwright suite.

## Workflow behavior

- Trigger: push to `main` and `dev`, plus `workflow_dispatch`.
- Backend: the runner's own checkout (current ref) is built and run by `frontend/scripts/start-backend.sh` via `E2E_BACKEND_DIR=${{ github.workspace }}`.
- Frontend: checked out to `frontend/` at `dev`; npm deps installed; Playwright browsers installed.
- DB: Postgres 16 service with `POSTGRES_DB=productdb_e2e`; schema created by backend migrations at startup; data truncated by Playwright `globalSetup`.
- Media/uploads: written to a temp `MEDIA_ROOT` removed by Playwright `globalTeardown` and the backend launcher.
- `psql` client is installed (needed by `start-backend.sh`).
- Artifacts uploaded on failure.

## Files
- New `.github/workflows/e2e.yml`
- New `decisions/2026-08-07.md` (decision record)
- New `plans/e2e-test.md` (this file)