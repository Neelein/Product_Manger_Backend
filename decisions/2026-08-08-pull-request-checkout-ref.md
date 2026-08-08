# Decision — Fix cross-repo checkout ref for pull_request events (2026-08-08)

## Problem
`cicd.yml` e2e job checks out the sibling repo (`Product_Manger_frontend`) with
`ref: ${{ github.ref_name }}`. For `pull_request` events `github.ref` is
`refs/pull/N/merge` and `github.ref_name` is `N/merge`, which is not a branch or
tag in the sibling repo. Result: git error "A branch or tag with the name
'25/merge' could not be found".

## Change
`ref: ${{ github.head_ref || github.ref_name }}`. On PR events `github.head_ref`
holds the source branch (which exists in both same-owner repos); on
push/workflow_dispatch events `github.head_ref` is empty and the fallback
`github.ref_name` (`main`/`dev`) is used.

## Why
Keeps the branch-matching behaviour for pushes while making PR runs resolvable.

## Alternatives rejected
- Guard the e2e cross-repo job with `if: github.event_name != 'pull_request'` —
  would silently skip E2E for PRs.
- `ref: ${{ github.event.pull_request.head.ref || github.ref_name }}` — same
  result as `github.head_ref` but more verbose.