# E2E checks branch-matching counterpart repo (2026-08-08)

## What
Changed the backend full-stack E2E workflow so it checks out the frontend at the same branch name as the workflow trigger (`${{ github.ref_name }}`) instead of the hardcoded `dev` ref.

## Why
Before, backend `main` merges were validated against `frontend@dev` rather than `frontend@main`, so the merged production pairing was not what was actually tested. Now a `dev` push tests dev code and a `main` merge tests main code.

## Changes
- `backend/.github/workflows/e2e.yml`: frontend checkout `ref: dev` → `ref: ${{ github.ref_name }}`. The backend under test is the repo's own workspace checkout, which already tracks the pushed branch.

## Notes
- Paired change in the frontend repo: `frontend/.github/workflows/e2e.yml` now checks out the backend at `${{ github.ref_name }}` too.
- `workflow_dispatch` uses the manually selected branch for `github.ref_name`.

## Verification
- YAML syntax validated.
- Pushed to `dev` to confirm the pipeline (main optional).