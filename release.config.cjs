"use strict";
module.exports = {
  branches: ["main"],
  plugins: [
    "@semantic-release/commit-analyzer",
    "@semantic-release/release-notes-generator",
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
