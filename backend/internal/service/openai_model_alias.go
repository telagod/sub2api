package service

import "strings"

// lastOpenAIModelSegment extracts the final path segment from a model
// identifier that may contain slashes (e.g. "org/model-name").
func lastOpenAIModelSegment(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return ""
	}
	if idx := strings.LastIndex(id, "/"); idx >= 0 {
		id = id[idx+1:]
	}
	return strings.TrimSpace(id)
}

// canonicalizeOpenAIModelAliasSpelling normalizes a model name by lower-casing,
// replacing underscores with hyphens, collapsing whitespace and double-hyphens,
// and applying known spelling corrections for GPT-5.x variants.
func canonicalizeOpenAIModelAliasSpelling(raw string) string {
	lowered := strings.ToLower(lastOpenAIModelSegment(raw))
	if lowered == "" {
		return ""
	}

	canonical := strings.ReplaceAll(lowered, "_", "-")
	canonical = strings.Join(strings.Fields(canonical), "-")
	for strings.Contains(canonical, "--") {
		canonical = strings.ReplaceAll(canonical, "--", "-")
	}

	if strings.HasPrefix(canonical, "gpt5") {
		canonical = "gpt-5" + strings.TrimPrefix(canonical, "gpt5")
	}
	if !strings.HasPrefix(canonical, "gpt-") && !strings.Contains(canonical, "codex") {
		return ""
	}

	corrections := []struct {
		from string
		to   string
	}{
		{"gpt-5.4mini", "gpt-5.4-mini"},
		{"gpt-5.4nano", "gpt-5.4-nano"},
		{"gpt-5.3-codexspark", "gpt-5.3-codex-spark"},
		{"gpt-5.3codexspark", "gpt-5.3-codex-spark"},
		{"gpt-5.3codex", "gpt-5.3-codex"},
	}
	for _, fix := range corrections {
		canonical = strings.ReplaceAll(canonical, fix.from, fix.to)
	}
	return canonical
}

// normalizeKnownOpenAICodexModel maps various user-provided model names
// to their canonical form using spelling canonicalization followed by
// prefix-based family matching.
func normalizeKnownOpenAICodexModel(raw string) string {
	canonical := canonicalizeOpenAIModelAliasSpelling(raw)
	if canonical == "" {
		return ""
	}

	if mapped := getNormalizedCodexModel(canonical); mapped != "" {
		return mapped
	}
	if strings.HasSuffix(canonical, "-openai-compact") {
		if mapped := getNormalizedCodexModel(strings.TrimSuffix(canonical, "-openai-compact")); mapped != "" {
			return mapped
		}
	}

	switch {
	case strings.Contains(canonical, "gpt-5.5"):
		return "gpt-5.5"
	case strings.Contains(canonical, "gpt-5.4-mini"):
		return "gpt-5.4-mini"
	case strings.Contains(canonical, "gpt-5.4-nano"):
		return "gpt-5.4-nano"
	case strings.Contains(canonical, "gpt-5.4"):
		return "gpt-5.4"
	case strings.Contains(canonical, "gpt-5.2"):
		return "gpt-5.2"
	case strings.Contains(canonical, "gpt-5.3-codex-spark"):
		return "gpt-5.3-codex-spark"
	case strings.Contains(canonical, "gpt-5.3-codex"):
		return "gpt-5.3-codex"
	case strings.Contains(canonical, "gpt-5.3"):
		return "gpt-5.3-codex"
	case strings.Contains(canonical, "codex"):
		return "gpt-5.3-codex"
	case strings.Contains(canonical, "gpt-5"):
		return "gpt-5.4"
	default:
		return ""
	}
}

// appendUsageBillingModelCandidate adds a model and its normalized variants
// to the candidate list, deduplicating by lowered name.
func appendUsageBillingModelCandidate(candidates []string, dedup map[string]struct{}, modelName string) []string {
	trimmed := strings.TrimSpace(modelName)
	if trimmed == "" {
		return candidates
	}
	insert := func(candidate string) {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			return
		}
		lo := strings.ToLower(candidate)
		if _, exists := dedup[lo]; exists {
			return
		}
		dedup[lo] = struct{}{}
		candidates = append(candidates, candidate)
	}

	insert(trimmed)
	if canonical := canonicalizeOpenAIModelAliasSpelling(trimmed); canonical != "" {
		insert(canonical)
	}
	if normalized := normalizeKnownOpenAICodexModel(trimmed); normalized != "" {
		insert(normalized)
	}
	return candidates
}

// usageBillingModelCandidates builds a deduplicated candidate list starting
// with the primary model, then alternates, each expanded with canonical and
// normalized forms.
func usageBillingModelCandidates(primary string, alternates ...string) []string {
	dedup := make(map[string]struct{}, 1+len(alternates))
	out := appendUsageBillingModelCandidate(nil, dedup, primary)
	for _, alt := range alternates {
		out = appendUsageBillingModelCandidate(out, dedup, alt)
	}
	return out
}

// firstUsageBillingModel returns the first non-empty candidate.
func firstUsageBillingModel(candidates []string) string {
	for _, c := range candidates {
		if trimmed := strings.TrimSpace(c); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
