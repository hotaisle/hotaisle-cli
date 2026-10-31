// Package skills contains the agent instructions shipped with the CLI binary.
package skills

import _ "embed"

// HotAisle is the same skill distributed in skills/hotaisle/SKILL.md.
//
//go:embed hotaisle/SKILL.md
var HotAisle string
