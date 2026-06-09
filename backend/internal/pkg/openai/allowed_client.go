package openai

import "strings"

const (
	// AllowedClientClaudeCode identifies the Claude Code CLI codex plugin.
	AllowedClientClaudeCode = "claude_code"
)

// AllowedClientEntry describes a permitted non-official Codex client signature.
// Originator must match exactly after normalization. UAContains is mandatory:
// an empty list or any blank marker causes a safe-fail (returns false) to
// prevent dual-factor matching from degrading to single-factor originator-only.
type AllowedClientEntry struct {
	Originator string
	UAContains []string
}

// Fixed registry of named preset client signatures.
var allowedClientRegistry = map[string]AllowedClientEntry{
	AllowedClientClaudeCode: {
		Originator: "Claude Code",
		UAContains: []string{"Claude Code/"},
	},
}

// IsAllowedClientMatch checks whether the request headers match a specific
// client entry. Both originator exact-match and all UA markers must be present.
func IsAllowedClientMatch(ua, orig string, entry AllowedClientEntry) bool {
	expectedOrig := normalizeCodexClientHeader(entry.Originator)
	if expectedOrig == "" {
		return false
	}
	if normalizeCodexClientHeader(orig) != expectedOrig {
		return false
	}
	if len(entry.UAContains) == 0 {
		return false
	}
	normalizedUA := normalizeCodexClientHeader(ua)
	for _, pattern := range entry.UAContains {
		normPattern := normalizeCodexClientHeader(pattern)
		if normPattern == "" {
			return false
		}
		if !strings.Contains(normalizedUA, normPattern) {
			return false
		}
	}
	return true
}

// MatchAllowedClients reports whether the request headers match any of the
// named preset client IDs. Unknown IDs are silently skipped; an empty list
// always returns false (default deny).
func MatchAllowedClients(ua, orig string, ids []string) bool {
	for _, clientID := range ids {
		preset, exists := allowedClientRegistry[normalizeCodexClientHeader(clientID)]
		if !exists {
			continue
		}
		if IsAllowedClientMatch(ua, orig, preset) {
			return true
		}
	}
	return false
}
