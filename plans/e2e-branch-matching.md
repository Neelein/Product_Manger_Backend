# Plan — E2E checks the branch-matching counterpart repo (2026-08-08)

## Goal
Make the full-stack E2E workflow check out the counterpart repository at the SAME branch name as the workflow trigger, so:
- push to `dev` → E2E tests `dev`-branch code
- merge to `main` → E2E tests `main`-branch code

## Changes
| File | Action | Description |
|------|--------|-------------|
| `backend/.github/workflows/e2e.yml` | Modify | Frontend checkout `ref: dev` → `ref: ${{ github.ref_name }}` (backend already uses its own workspace code, which is branch-correct) |

## Result matrix
| Trigger | Tested combo |
|---------|--------------|
| backend push `dev` | backend dev (own) + frontend dev |
| backend push `main` (post-merge) | backend main (own) + frontend main |

## Verification
- YAML syntax validation.
- Push to `dev` and watch the GitHub Action run (main optional).