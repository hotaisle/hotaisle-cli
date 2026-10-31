import assert from "node:assert/strict";
import { test } from "node:test";
import { analyzeCommits } from "@semantic-release/commit-analyzer";
import { generateNotes } from "@semantic-release/release-notes-generator";
import config from "./release.config.cjs";

const [, analyzer] = config.plugins.find(
  ([name]) => name === "@semantic-release/commit-analyzer"
);
const [, notesConfig] = config.plugins.find(
  ([name]) => name === "@semantic-release/release-notes-generator"
);
const logger = { log: () => undefined };

for (const [message, expected] of [
  ["feat(cli): add a command", "minor"],
  ["fix(cli): correct a command", "patch"],
  ["docs: explain a command", null],
  ["feat(cli)!: change the output default", "major"],
  ["fix!: change the output default", "major"],
  [
    "fix(cli): change the output default\n\nBREAKING CHANGE: Add --json in scripts.",
    "major",
  ],
  [
    "fix(cli): change the output default\n\nBREAKING CHANGES: Add --json in scripts.",
    "major",
  ],
]) {
  for (const version of ["0.10.8", "1.2.3"]) {
    test(`release analysis from ${version}: ${message.split("\n")[0]}`, async () => {
      const release = await analyzeCommits(analyzer, {
        commits: [{ hash: "a".repeat(40), message }],
        cwd: process.cwd(),
        lastRelease: { version },
        logger,
      });
      assert.equal(release, expected);
    });
  }
}

test("release notes prominently include breaking defaults and migration", async () => {
  const notes = await generateNotes(notesConfig, {
    commits: [
      {
        hash: "a".repeat(40),
        message:
          "feat(cli)!: default to human-readable output\n\nBREAKING CHANGE: Default command output is human-readable text. Add --json before commands in scripts; use --agent for Markdown. See CHANGELOG.md.",
      },
    ],
    cwd: process.cwd(),
    lastRelease: { gitTag: "v1.2.3" },
    nextRelease: { gitTag: "v2.0.0", version: "2.0.0" },
    options: { repositoryUrl: "https://github.com/hotaisle/hotaisle-cli.git" },
  });
  for (const text of [
    "BREAKING CHANGES",
    "Features",
    "default to human-readable output",
    "Add --json before commands in scripts",
    "use --agent for Markdown",
    "CHANGELOG.md",
  ]) {
    assert.ok(notes.includes(text), `Missing release note: ${text}`);
  }
});

test("bang-only changes also appear in the breaking-changes section", async () => {
  const notes = await generateNotes(notesConfig, {
    commits: [
      { hash: "b".repeat(40), message: "fix(cli)!: change the output default" },
    ],
    cwd: process.cwd(),
    lastRelease: { gitTag: "v1.2.3" },
    nextRelease: { gitTag: "v2.0.0", version: "2.0.0" },
    options: { repositoryUrl: "https://github.com/hotaisle/hotaisle-cli.git" },
  });
  assert.ok(notes.includes("BREAKING CHANGES"));
  assert.ok(notes.includes("Bug Fixes"));
  assert.ok(notes.includes("change the output default"));
});
