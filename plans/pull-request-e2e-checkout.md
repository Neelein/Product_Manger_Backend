# Plan — Fix cross-repo checkout on pull_request events (2026-08-08)

## Goal
The e2e job of `cicd.yml` checks out the other repo with `ref: ${{ github.ref_name }}`. On `pull_request` events `github.ref_name` resolves to `25/merge` (not a branch/tag name), so the cross-repo checkout fails with "A branch or tag with the name '25/merge' could not be found". Use the PR source branch when available.

## Changes
| File | Action | Description |
|------|--------|-------------|
| `.github/workflows/cicd.yml` | Modify | `ref: ${{ github.head_ref || github.ref_name }}` — PR events use the head branch; push/workflow_dispatch fall back to `github.ref_name` |
| `decisions/2026-08-08-pull-request-checkout-ref.md` | New | Decision record |

## Verification
- YAML parses (ruby psych).
- Push-based runs (ref_name = dev/main) behave identically to before.
- A pull_request targeting main now checks out the source branch instead of `refs/pull/N/merge`.