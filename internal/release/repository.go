package release

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// PlanRepository only reads Git history. Tags outside HEAD's ancestry and
// prerelease tags do not affect the next stable version.
func PlanRepository(ctx context.Context, dir, repository string) (Plan, error) {
	head, err := git(ctx, dir, "rev-parse", "--verify", "HEAD")
	if err != nil {
		return Plan{}, err
	}
	tags, err := git(ctx, dir, "tag", "--merged", "HEAD", "--list", "v*")
	if err != nil {
		return Plan{}, err
	}
	previous := LatestTag(strings.Fields(tags))
	revisions := "HEAD"
	if previous != "" {
		revisions = previous + "..HEAD"
	}
	log, err := git(ctx, dir, "log", "--topo-order", "-z", "--format=%H%x00%B", revisions, "--")
	if err != nil {
		return Plan{}, err
	}
	var commits []Commit
	for log != "" {
		hash, rest, ok := strings.Cut(log, "\x00")
		if !ok {
			return Plan{}, fmt.Errorf("invalid Git log record")
		}
		message, rest, ok := strings.Cut(rest, "\x00")
		if !ok {
			return Plan{}, fmt.Errorf("incomplete Git log record")
		}
		commits = append(commits, Commit{Hash: hash, Message: message})
		log = rest
	}
	return CreatePlan(repository, previous, strings.TrimSpace(head), commits)
}

func git(ctx context.Context, dir string, args ...string) (string, error) {
	command := exec.CommandContext(ctx, "git", args...)
	command.Dir = dir
	command.WaitDelay = 5 * time.Second
	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w", args[0], err)
	}
	return string(output), nil
}
