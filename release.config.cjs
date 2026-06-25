module.exports = {
  branches: ["main"],
  tagFormat: "v${version}",
  plugins: [
    "@semantic-release/commit-analyzer",
    "@semantic-release/release-notes-generator",
    [
      "@semantic-release/exec",
      {
        prepareCmd: "VERSION=${nextRelease.gitTag} just release",
        successCmd:
          'echo "released=true" >> "$GITHUB_OUTPUT" && echo "new_tag=${nextRelease.gitTag}" >> "$GITHUB_OUTPUT"',
      },
    ],
    [
      "@semantic-release/github",
      {
        assets: ["dist/*.zip", "dist/*.tar.gz", "dist-pkg/*"],
        successComment: false,
        failComment: false,
        labels: false,
      },
    ],
  ],
};
