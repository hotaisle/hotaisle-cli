package release_test

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/hotaisle/hotaisle-cli/internal/release"
)

const head = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestReleasePolicy(t *testing.T) {
	for _, previous := range []string{"v0.10.8", "v1.2.3"} {
		for _, tc := range []struct {
			message string
			index   int
		}{
			{"fix(cli): correct a command", 2},
			{"feat(cli): add a command", 1},
			{"docs: explain a command", 2},
			{"chore(deps): refresh modules", 2},
			{"a non-conventional commit", 2},
			{"feat(cli)!: change defaults", 0},
			{"fix!: change defaults", 0},
			{"fix(cli): change defaults\n\nBREAKING CHANGE: Add --json in scripts.", 0},
			{"fix(cli): change defaults\n\nBREAKING CHANGES: Add --json in scripts.", 0},
			{"fix(cli): explain BREAKING CHANGE: in prose", 2},
		} {
			t.Run(previous+"/"+strings.Split(tc.message, "\n")[0], func(t *testing.T) {
				plan, err := release.CreatePlan("hotaisle/hotaisle-cli", previous, head, []release.Commit{{Hash: head, Message: tc.message}})
				if err != nil {
					t.Fatal(err)
				}
				want := []string{"v1.0.0", "v0.11.0", "v0.10.9"}[tc.index]
				if previous == "v1.2.3" {
					want = []string{"v2.0.0", "v1.3.0", "v1.2.4"}[tc.index]
				}
				if plan.Tag != want {
					t.Fatalf("tag = %q, want %q", plan.Tag, want)
				}
			})
		}
	}
}

func TestReleaseNotesAndHighestBump(t *testing.T) {
	commits := []release.Commit{
		{Hash: head, Message: "feat(cli)!: default to readable text (#79)\n\nBREAKING CHANGE: Add --json before commands in scripts.\n\nUse --agent for Markdown. See CHANGELOG.md.\nRefs: #78\n\nBREAKING CHANGES: A second migration."},
		{Hash: strings.Repeat("b", 40), Message: "fix(cli): retain zero values"},
		{Hash: strings.Repeat("c", 40), Message: "docs: explain setup"},
		{Hash: strings.Repeat("d", 40), Message: "feat(keys): add key registration"},
		{Hash: strings.Repeat("e", 40), Message: "unconventional subject"},
	}
	plan, err := release.CreatePlan("hotaisle/hotaisle-cli", "v0.10.8", head, commits)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Tag != "v1.0.0" {
		t.Fatal(plan.Tag)
	}
	for _, text := range []string{
		"compare/v0.10.8...v1.0.0", "### BREAKING CHANGES", "Add --json before commands in scripts.",
		"Use --agent for Markdown. See CHANGELOG.md.", "A second migration.", "### Features", "### Bug Fixes",
		"### Documentation", "### Other Changes", "**cli:** default to readable text", "[aaaaaaa]",
		"[#79](https://github.com/hotaisle/hotaisle-cli/issues/79)",
	} {
		if !strings.Contains(plan.Notes, text) {
			t.Errorf("missing %q in notes:\n%s", text, plan.Notes)
		}
	}
	if strings.Contains(plan.Notes, "Refs:") {
		t.Fatal("unrelated footer included in migration instructions")
	}
	if strings.Index(plan.Notes, "BREAKING CHANGES") > strings.Index(plan.Notes, "### Features") {
		t.Fatal("migration instructions must appear before change summaries")
	}
}

func TestBangOnlyAndMultilineNotes(t *testing.T) {
	for _, message := range []string{
		"fix(cli)!: change the default",
		"fix(cli): change the default\r\n\r\nBREAKING CHANGE:\r\n  Add --json.\r\n\r\n  Keep scripts working.",
	} {
		plan, err := release.CreatePlan("hotaisle/hotaisle-cli", "v1.2.3", head, []release.Commit{{Hash: head, Message: message}})
		if err != nil || plan.Tag != "v2.0.0" || !strings.Contains(plan.Notes, "### BREAKING CHANGES") {
			t.Fatalf("plan = %+v, error = %v", plan, err)
		}
	}
}

func TestInitialAndEmptyReleases(t *testing.T) {
	plan, err := release.CreatePlan("hotaisle/hotaisle-cli", "", head, []release.Commit{{Hash: head, Message: "fix: initial change"}})
	if err != nil || plan.Tag != "v1.0.0" || !strings.HasPrefix(plan.Notes, "## v1.0.0\n") {
		t.Fatalf("plan = %+v, error = %v", plan, err)
	}
	plan, err = release.CreatePlan("hotaisle/hotaisle-cli", "v1.0.0", head, nil)
	if err != nil || plan.Tag != "" || plan.Notes != "" {
		t.Fatalf("empty plan = %+v, error = %v", plan, err)
	}
	var out strings.Builder
	if err := plan.WriteOutputs(&out); err != nil {
		t.Fatal(err)
	}
	if out.String() != "release_needed=false\nnew_tag=\ncommit="+head+"\n" {
		t.Fatal(out.String())
	}
}

func TestRevertedChanges(t *testing.T) {
	feature := release.Commit{Hash: strings.Repeat("b", 40), Message: "feat(cli)!: change defaults"}
	revert := release.Commit{Hash: strings.Repeat("c", 40), Message: "Revert \"feat(cli)!: change defaults\"\n\nThis reverts commit " + feature.Hash + "."}
	restore := release.Commit{Hash: strings.Repeat("d", 40), Message: "revert: restore defaults change\n\nThis reverts commit " + revert.Hash + "."}
	for _, tc := range []struct {
		name    string
		commits []release.Commit
		tag     string
	}{
		{"cancel breaking change", []release.Commit{revert, feature}, "v1.2.4"},
		{"restore breaking change", []release.Commit{restore, revert, feature}, "v2.0.0"},
		{"revert an earlier release", []release.Commit{revert}, "v1.2.4"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			plan, err := release.CreatePlan("hotaisle/hotaisle-cli", "v1.2.3", head, tc.commits)
			if err != nil || plan.Tag != tc.tag {
				t.Fatalf("plan = %+v, error = %v", plan, err)
			}
			if tc.tag == "v1.2.4" && strings.Contains(plan.Notes, "BREAKING CHANGES") {
				t.Fatal("reverted migrations should not appear as new breaking changes")
			}
		})
	}
}

func TestLatestStableTag(t *testing.T) {
	tags := []string{"v0.10.9", "v1.9.9", "v1.10.0", "v1.2.99", "v99.0.0-beta.1", "v01.99.0", "not-a-version", "v1.99.0+build"}
	if got := release.LatestTag(tags); got != "v1.10.0" {
		t.Fatal(got)
	}
	if got := release.LatestTag([]string{"v1.0.0-rc.1", "unrelated"}); got != "" {
		t.Fatal(got)
	}
}

func TestRejectInvalidPlanInputs(t *testing.T) {
	for _, tc := range []struct{ repo, tag, sha, message string }{
		{"../owner/repo", "v1.0.0", head, "fix: change"},
		{"owner/repo\ninjected=true", "v1.0.0", head, "fix: change"},
		{"owner/repo", "v1.0.0-beta.1", head, "fix: change"},
		{"owner/repo", "v1.0.0", "HEAD", "fix: change"},
		{"owner/repo", "v1.0.0", head, " "},
		{"owner/repo", "v1.0.18446744073709551615", head, "fix: change"},
	} {
		_, err := release.CreatePlan(tc.repo, tc.tag, tc.sha, []release.Commit{{Hash: head, Message: tc.message}})
		if err == nil {
			t.Errorf("accepted %+v", tc)
		}
	}
}

type failedWriter struct{}

func (failedWriter) Write([]byte) (int, error) { return 0, errors.New("output failed") }

func TestOutputFailure(t *testing.T) {
	if err := (release.Plan{}).WriteOutputs(failedWriter{}); err == nil {
		t.Fatal("output errors must be reported")
	}
}

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	command := exec.CommandContext(t.Context(), "git", args...)
	command.Dir = dir
	command.WaitDelay = 5 * time.Second
	out, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func commit(t *testing.T, dir, message string) string {
	t.Helper()
	branch := git(t, dir, "branch", "--show-current")
	if !strings.HasPrefix(branch, "test/") {
		t.Fatalf("refusing to commit on %q", branch)
	}
	git(t, dir, "commit", "--allow-empty", "--no-gpg-sign", "-m", message)
	return git(t, dir, "rev-parse", "HEAD")
}

func TestGitHistoryPlanningDoesNotMutateRepository(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "--initial-branch=main")
	git(t, dir, "switch", "-c", "test/release-history")
	for key, value := range map[string]string{"user.name": "Release Tests", "user.email": "tests@example.invalid", "core.hooksPath": t.TempDir(), "tag.gpgsign": "false"} {
		git(t, dir, "config", key, value)
	}
	commit(t, dir, "fix: initial change")
	initial, err := release.PlanRepository(t.Context(), dir, "hotaisle/hotaisle-cli")
	if err != nil || initial.Tag != "v1.0.0" {
		t.Fatalf("initial plan = %+v, error = %v", initial, err)
	}
	git(t, dir, "tag", "-a", "v0.10.8", "-m", "Previous release")
	git(t, dir, "switch", "-c", "test/unmerged")
	commit(t, dir, "feat!: unreleased work on another branch")
	git(t, dir, "tag", "v99.0.0")
	git(t, dir, "switch", "test/release-history")
	sha := commit(t, dir, "feat(cli)!: change defaults\n\nA body with multiple paragraphs.\n\nBREAKING CHANGE: Add --json.\n\nKeep migration details intact.")
	git(t, dir, "tag", "v9.0.0-rc.1")
	before := git(t, dir, "show-ref")
	plan, err := release.PlanRepository(t.Context(), dir, "hotaisle/hotaisle-cli")
	if err != nil {
		t.Fatal(err)
	}
	if plan.PreviousTag != "v0.10.8" || plan.Tag != "v1.0.0" || plan.Commit != sha || !strings.Contains(plan.Notes, "Keep migration details intact.") {
		t.Fatalf("unexpected plan: %+v", plan)
	}
	var out strings.Builder
	if err := plan.WriteOutputs(&out); err != nil {
		t.Fatal(err)
	}
	if want := fmt.Sprintf("release_needed=true\nnew_tag=v1.0.0\ncommit=%s\n", sha); out.String() != want {
		t.Fatal(out.String())
	}
	if git(t, dir, "show-ref") != before || git(t, dir, "status", "--porcelain") != "" {
		t.Fatal("release planning mutated the repository")
	}
	git(t, dir, "tag", "v1.0.0")
	plan, err = release.PlanRepository(t.Context(), dir, "hotaisle/hotaisle-cli")
	if err != nil || plan.Tag != "" {
		t.Fatalf("already released plan = %+v, error = %v", plan, err)
	}
	commit(t, dir, "docs: update instructions")
	plan, err = release.PlanRepository(t.Context(), dir, "hotaisle/hotaisle-cli")
	if err != nil || plan.Tag != "v1.0.1" {
		t.Fatalf("documentation plan = %+v, error = %v", plan, err)
	}
}

func TestInvalidRepository(t *testing.T) {
	if _, err := release.PlanRepository(t.Context(), t.TempDir(), "hotaisle/hotaisle-cli"); err == nil {
		t.Fatal("a directory without Git history must fail")
	}
}
