## Agent rules

- **Every change must have tests.** No code is merged without corresponding tests.
- **Do not design or decide alone.** Every design decision must be discussed with me first.
- **Code format must follow the skill.** Load the relevant skill (e.g. `golang-code-style`, `golang-testing`) before writing or reviewing code.
- **Do not write production code.** Use a sub agent (Task tool) to generate all code. You only review and orchestrate.
- **Do not push to GitHub without approval.** Always ask for explicit confirmation before any `git push`.
- **Document new features in decisions.** When creating a decision file under `decisions/`, write a brief English description of the decision made for each new feature.
