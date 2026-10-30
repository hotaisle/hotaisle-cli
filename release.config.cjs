"use strict";
const parserOpts = {
  breakingHeaderPattern:
    /^(?<type>\w+)(?:\((?<scope>[^)]+)\))?!: (?<subject>.+)$/u,
  headerCorrespondence: ["type", "scope", "subject"],
  headerPattern: /^(?<type>\w+)(?:\((?<scope>[^)]+)\))?!?: (?<subject>.+)$/u,
  noteKeywords: ["BREAKING CHANGE", "BREAKING CHANGES"],
};

module.exports = {
  branches: ["main"],
  plugins: [
    ["@semantic-release/commit-analyzer", { parserOpts }],
    [
      "@semantic-release/release-notes-generator",
      {
        parserOpts,
        writerOpts: {
          transform(commit, context) {
            const sections = {
              build: "Build System",
              chore: "Maintenance",
              ci: "Continuous Integration",
              docs: "Documentation",
              feat: "Features",
              fix: "Bug Fixes",
              perf: "Performance",
              refactor: "Refactoring",
              revert: "Reverts",
              style: "Styles",
              test: "Tests",
            };
            const subjectIssues = new Set(commit.subject?.match(/#\d+/gu));
            return {
              ...commit,
              notes: commit.notes.map((note) => ({
                ...note,
                title: "BREAKING CHANGES",
              })),
              references: commit.references.filter(
                (reference) => !subjectIssues.has(`#${reference.issue}`)
              ),
              scope: commit.scope === "*" ? "" : commit.scope,
              shortHash: commit.hash.slice(0, 7),
              subject: commit.subject?.replace(
                /#(?<issue>\d+)/gu,
                `[#$1](${context.host}/${context.owner}/${context.repository}/issues/$1)`
              ),
              type: Object.hasOwn(sections, commit.type)
                ? sections[commit.type]
                : commit.type,
            };
          },
        },
      },
    ],
    [
      "@semantic-release/exec",
      {
        analyzeCommitsCmd: "test ${commits.length} -gt 0 && echo patch || true",
        prepareCmd: "VERSION=${nextRelease.gitTag} just release",
        successCmd:
          'echo "released=true" >> "$GITHUB_OUTPUT" && echo "new_tag=${nextRelease.gitTag}" >> "$GITHUB_OUTPUT"',
      },
    ],
    [
      "@semantic-release/github",
      {
        assets: ["dist/*.zip", "dist/*.tar.gz", "dist-pkg/*"],
        failComment: false,
        labels: false,
        successComment: false,
      },
    ],
  ],
  tagFormat: "v${version}",
};
