// Package release plans releases without creating tags or contacting GitHub.
package release

import (
	"cmp"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var (
	versionPattern  = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)
	headerPattern   = regexp.MustCompile(`^(\w+)(?:\(([^()]+)\))?(!)?: (.+)$`)
	breakingPattern = regexp.MustCompile(`^BREAKING CHANGES?:[\t ]*(.*)$`)
	footerPattern   = regexp.MustCompile(`^(?:[\w-]+:[\t ]|(?:Closes|Fixes|Refs|Resolves) #)`)
	revertPattern   = regexp.MustCompile(`(?m)^This reverts commit ([a-f0-9]{40}|[a-f0-9]{64})\.$`)
	issuePattern    = regexp.MustCompile(`#([0-9]+)`)
	hashPattern     = regexp.MustCompile(`^(?:[a-f0-9]{40}|[a-f0-9]{64})$`)
	repoPattern     = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)
)

// Commit is a complete commit message, including any migration footers.
type Commit struct {
	Hash    string
	Message string
}

// Plan describes the next release. An empty Tag means there are no new commits.
type Plan struct {
	PreviousTag string
	Tag         string
	Commit      string
	Notes       string
}

type version [3]uint64

func parseVersion(tag string) (version, bool) {
	parts := versionPattern.FindStringSubmatch(tag)
	if parts == nil {
		return version{}, false
	}
	var result version
	for i := range result {
		n, err := strconv.ParseUint(parts[i+1], 10, 64)
		if err != nil {
			return version{}, false
		}
		result[i] = n
	}
	return result, true
}

// LatestTag selects the highest stable vMAJOR.MINOR.PATCH tag from reachable tags.
func LatestTag(tags []string) string {
	var latest string
	var highest version
	for _, tag := range tags {
		v, ok := parseVersion(tag)
		if !ok {
			continue
		}
		if latest == "" || compare(v, highest) > 0 {
			latest, highest = tag, v
		}
	}
	return latest
}

func compare(a, b version) int {
	for i := range a {
		if n := cmp.Compare(a[i], b[i]); n != 0 {
			return n
		}
	}
	return 0
}

type change struct {
	Commit
	kind, scope, subject string
	breaking             bool
	notes                []string
}

func parseCommit(commit Commit) change {
	lines := strings.Split(strings.ReplaceAll(strings.TrimSpace(commit.Message), "\r\n", "\n"), "\n")
	c := change{Commit: commit, subject: lines[0]}
	if parts := headerPattern.FindStringSubmatch(c.subject); parts != nil {
		c.kind, c.scope, c.subject = parts[1], parts[2], parts[4]
		c.breaking = parts[3] == "!"
	}
	for i := 1; i < len(lines); i++ {
		match := breakingPattern.FindStringSubmatch(lines[i])
		if match == nil {
			continue
		}
		c.breaking = true
		note := []string{match[1]}
		for i+1 < len(lines) && !breakingPattern.MatchString(lines[i+1]) && !footerPattern.MatchString(lines[i+1]) {
			i++
			note = append(note, lines[i])
		}
		if text := strings.TrimSpace(strings.Join(note, "\n")); text != "" {
			c.notes = append(c.notes, text)
		}
	}
	if c.breaking && len(c.notes) == 0 {
		c.notes = []string{c.subject}
	}
	return c
}

// CreatePlan retains the release policy: every nonempty change set releases at
// least a patch; features select minor; breaking changes select major, even at 0.x.
// A repository without a stable release tag starts at v1.0.0.
// Commits must be ordered newest first, as in Git's topological log order.
func CreatePlan(repository, previousTag, head string, commits []Commit) (Plan, error) {
	plan := Plan{PreviousTag: previousTag, Commit: head}
	if !repoPattern.MatchString(repository) || !hashPattern.MatchString(head) {
		return Plan{}, fmt.Errorf("invalid repository or HEAD commit")
	}
	v, ok := parseVersion(previousTag)
	if previousTag != "" && !ok {
		return Plan{}, fmt.Errorf("invalid stable release tag %q", previousTag)
	}
	if len(commits) == 0 {
		return plan, nil
	}
	changes := make([]change, 0, len(commits))
	bump := 2 // patch, including documentation and maintenance commits
	for _, commit := range commits {
		if !hashPattern.MatchString(commit.Hash) || strings.TrimSpace(commit.Message) == "" {
			return Plan{}, fmt.Errorf("invalid commit in release history")
		}
		changes = append(changes, parseCommit(commit))
	}
	changes = withoutRevertedPairs(changes)
	for _, c := range changes {
		if c.breaking {
			bump = 0
		} else if c.kind == "feat" {
			bump = min(bump, 1)
		}
	}
	if previousTag == "" {
		v = version{1, 0, 0}
	} else {
		if v[bump] == math.MaxUint64 {
			return Plan{}, fmt.Errorf("release version overflow")
		}
		v[bump]++
		for i := bump + 1; i < len(v); i++ {
			v[i] = 0
		}
	}
	plan.Tag = fmt.Sprintf("v%d.%d.%d", v[0], v[1], v[2])
	plan.Notes = renderNotes(repository, plan, changes)
	return plan, nil
}

// Git supplies newest commits first, so a revert of a revert restores the
// original change. A revert targeting an earlier release remains a patch.
func withoutRevertedPairs(changes []change) []change {
	byHash := make(map[string]int, len(changes))
	for i, c := range changes {
		byHash[c.Hash] = i
	}
	removed := make([]bool, len(changes))
	for i, c := range changes {
		if removed[i] || (c.kind != "revert" && !strings.HasPrefix(c.subject, `Revert "`)) {
			continue
		}
		changes[i].kind = "revert"
		if match := revertPattern.FindStringSubmatch(c.Message); match != nil {
			if target, exists := byHash[match[1]]; exists && target > i {
				removed[i], removed[target] = true, true
			}
		}
	}
	kept := make([]change, 0, len(changes))
	for i, c := range changes {
		if !removed[i] {
			kept = append(kept, c)
		}
	}
	return kept
}

func renderNotes(repository string, plan Plan, changes []change) string {
	base := "https://github.com/" + repository
	var out strings.Builder
	if plan.PreviousTag == "" {
		fmt.Fprintf(&out, "## %s\n", plan.Tag)
	} else {
		fmt.Fprintf(&out, "## [%s](%s/compare/%s...%s)\n", plan.Tag, base, plan.PreviousTag, plan.Tag)
	}
	breaking := false
	for _, c := range changes {
		for _, note := range c.notes {
			if !breaking {
				out.WriteString("\n### BREAKING CHANGES\n\n")
				breaking = true
			}
			writeBullet(&out, c.scope, note)
		}
	}
	sections := []struct{ kind, title string }{
		{"feat", "Features"},
		{"fix", "Bug Fixes"},
		{"perf", "Performance"},
		{"revert", "Reverts"},
		{"docs", "Documentation"},
		{"refactor", "Refactoring"},
		{"build", "Build System"},
		{"ci", "Continuous Integration"},
		{"chore", "Maintenance"},
		{"style", "Styles"},
		{"test", "Tests"},
		{"", "Other Changes"},
	}
	for _, section := range sections {
		started := false
		for _, c := range changes {
			kind := c.kind
			switch kind {
			case "feat", "fix", "perf", "revert", "docs", "refactor", "build", "ci", "chore", "style", "test":
			default:
				kind = ""
			}
			if kind != section.kind {
				continue
			}
			if !started {
				fmt.Fprintf(&out, "\n### %s\n\n", section.title)
				started = true
			}
			subject := issuePattern.ReplaceAllString(c.subject, "[#$1]("+base+"/issues/$1)")
			writeBullet(&out, c.scope, fmt.Sprintf("%s ([%s](%s/commit/%s))", subject, c.Hash[:7], base, c.Hash))
		}
	}
	return out.String()
}

func writeBullet(out *strings.Builder, scope, text string) {
	out.WriteString("* ")
	if scope != "" && scope != "*" {
		fmt.Fprintf(out, "**%s:** ", scope)
	}
	first := true
	for line := range strings.SplitSeq(text, "\n") {
		if !first {
			out.WriteString("  ")
		}
		out.WriteString(line)
		out.WriteByte('\n')
		first = false
	}
}

// WriteOutputs writes GitHub Actions outputs, without claiming publication.
func (p Plan) WriteOutputs(out io.Writer) error {
	_, err := fmt.Fprintf(out, "release_needed=%t\nnew_tag=%s\ncommit=%s\n", p.Tag != "", p.Tag, p.Commit)
	return err
}
