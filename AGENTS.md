# AGENTS.md

- Always follow the instructions in this `AGENTS.md` file.
- Always use Conventional Commits for commit messages.
- Never include Codex, Claude, or other AI-assistant branding in branch names, commit messages, PR titles, PR bodies, or other generated labels.
- Prefer the "don't repeat yourself" (DRY) code that keeps the result clear and maintainable.
- Test packages should use the external `foo_test` package form unless there is a specific reason to test unexported internals.
- Run `just test`, `just link`, `just fmt after making code changes and then clean up any issues.
- Run all `git` commands sequentially and not parallel.
