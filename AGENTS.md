## Agent rules

- **Every change must have tests.** No code is merged without corresponding tests.
- **Backend changes must pass E2E tests.** Any backend change that alters user-facing behavior (API, auth, session, media) must be validated with the Playwright E2E suite before merge. Run locally via `frontend/scripts/start-backend.sh` + `npx playwright test`, or rely on the `.github/workflows/e2e.yml` CI workflow.
- **Do not design or decide alone.** Every design decision must be discussed with me first.
- **Code format must follow the skill.** Load the relevant skill (e.g. `golang-code-style`, `golang-testing`) before writing or reviewing code.
- **Do not write production code.** Use a sub agent (Task tool) to generate all code. You only review and orchestrate.
- **Do not push to GitHub without approval.** Always ask for explicit confirmation before any `git push`.
- **Discuss every design decision and modification first.** Every design decision and modification must be discussed with me before implementation.
- **Record plans in the shared docs workflow.** A plan may be recorded in the docs JSON/date records before the change.
- **Follow the confirmed plan.** Implementation must follow the confirmed plan.
- **Update and verify the shared change log for every change.** Every code, API, database, migration, configuration, test, or documentation modification must update `../docs/data/entries.json` and run `npm run generate && npm test` from `../docs`.
- **Shared change log workflow.** Every modification, including code, API, database, migration, configuration, test, or documentation changes, must synchronously update `../docs/data/entries.json`.
- **Shared change log verification.** Before completion, run `cd ../docs && npm run generate && npm test`, then confirm that `../docs/index.html` and the relevant `../docs/dates/YYYY-MM-DD.html` exist and contain the change.
- **Documentation changes are logged too.** A documentation-only change must also leave a record in `../docs/data/entries.json`; do not edit generated index or date pages directly.
- **Shared log replaces Markdown records for this workflow.** Do not create or restore `plans/` or `decisions/` Markdown files for these changes; use the JSON record with the fixed categories `決定`, `新增 Feature`, and `修改 Error / Bug`.
