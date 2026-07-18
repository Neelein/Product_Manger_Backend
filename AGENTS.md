## Agent rules

- **Every change must have tests.** No code is merged without corresponding tests.
- **Do not design or decide alone.** Every design decision must be discussed with me first.
- **Code format must follow the skill.** Load the relevant skill (e.g. `golang-code-style`, `golang-testing`) before writing or reviewing code.
- **Do not write production code.** Use a sub agent (Task tool) to generate all code. You only review and orchestrate.
- **Do not push to GitHub without approval.** Always ask for explicit confirmation before any `git push`.
- **Every change requires a decisions file.** Whenever you make any code change (feature, refactor, bug fix, migration, config), document it in a decision file under `decisions/`. Write a brief English description of what was changed and why. Decision files use date-based names (e.g. `2026-07-18.md`).
- **If not using plan mode, verify plan file first.** If we did not discuss using plan mode, before making any changes, confirm whether the plan file needs to be updated. If it does, update the plan file first, then start implementation.
- **Plan before implementation.** After discussion and before writing any code, create a plan file under `plans/` and a decisions file under `decisions/`. Implement strictly according to the plan. Plan files use topic-based names (e.g. `inventory-system.md`), decisions use date-based names (e.g. `2026-07-18.md`).
