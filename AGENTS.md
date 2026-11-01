# AGENTS.md

- Always follow the instructions in this `AGENTS.md` file.
- Always use Conventional Commits for commit messages.
- Changes to existing defaults or output contracts must use a breaking Conventional Commit (`type(scope)!: subject`) with a `BREAKING CHANGE:` footer explaining the migration. Add the user-visible change and migration steps to `CHANGELOG.md`; keep the footer useful in generated release notes.
- Never commit directly to `main`, `master`, or any repository's default branch. Always create or switch to a descriptive non-default branch first, including for `commit`, `add and commit`, `commit all`, and `release` requests.
- Before staging changes or creating a commit, run `git branch --show-current` and verify the result is the intended non-default branch. If on a default branch or detached HEAD, create and switch to a branch before proceeding.
- Never include Codex, Claude, or other AI-assistant branding in branch names, commit messages, PR titles, PR bodies, or other generated labels.
- Prefer the "don't repeat yourself" (DRY) code that keeps the result clear and maintainable.
- Test packages should use the external `foo_test` package form unless there is a specific reason to test unexported internals.
- Treat every Go LSP warning or hint as an error and fix it; never suppress diagnostics.
- Run `just workflow-check` for GitHub Actions changes and fix every actionlint diagnostic.
- Run `just test`, `just fmt` after making code changes and then clean up any issues.
- Run all `git` commands sequentially and not parallel.
