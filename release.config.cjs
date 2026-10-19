"use strict";
module.exports = {
  branches: ["main"],
  plugins: [
    "@semantic-release/commit-analyzer",
    [
      "@semantic-release/release-notes-generator",
      {
        preset: "conventionalcommits",
        presetConfig: {
          types: [
            { hidden: false, section: "Features", type: "feat" },
            { hidden: false, section: "Bug Fixes", type: "fix" },
            { hidden: false, section: "Performance", type: "perf" },
            { hidden: false, section: "Reverts", type: "revert" },
            { hidden: false, section: "Documentation", type: "docs" },
            { hidden: false, section: "Styles", type: "style" },
            { hidden: false, section: "Maintenance", type: "chore" },
            { hidden: false, section: "Refactoring", type: "refactor" },
            { hidden: false, section: "Tests", type: "test" },
            { hidden: false, section: "Build System", type: "build" },
            { hidden: false, section: "Continuous Integration", type: "ci" },
          ],
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
