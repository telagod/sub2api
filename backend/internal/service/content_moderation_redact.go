package service

import (
	"regexp"
	"strings"
)

var contentModerationSecretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bhttps?://[^\s"'<>，。；、]+`),
	regexp.MustCompile(`(?i)\b((?:api[_-]?key|apikey|access[_-]?token|refresh[_-]?token|id[_-]?token|session[_-]?token|token|session|cookie|set[_-]?cookie|authorization|bearer|password|passwd|pwd|secret|client[_-]?secret|private[_-]?key)\s*[:=]\s*)(["']?)[^"'\s,;，。；、]{6,}`),
	regexp.MustCompile(`(?i)\b(Bearer\s+)[A-Za-z0-9._~+/=-]{12,}`),
	regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\b`),
	regexp.MustCompile(`(?i)\b(?:sk|sk-proj|sk-ant|sess|rk|pk|ak|api|key|token|secret)[_-][A-Za-z0-9._~+/=-]{12,}\b`),
	regexp.MustCompile(`\b[0-9a-fA-F]{32,}\b`),
	regexp.MustCompile(`\b[A-Za-z0-9_-]{48,}\b`),
	regexp.MustCompile(`\b[A-Za-z0-9+/]{48,}={0,2}\b`),
	regexp.MustCompile(`\b[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}\b`),
}

func redactContentModerationSecrets(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return ""
	}
	result := trimmed
	for i, pat := range contentModerationSecretPatterns {
		switch i {
		case 1:
			result = pat.ReplaceAllString(result, `${1}${2}[REDACTED]`)
		case 2:
			result = pat.ReplaceAllString(result, `${1}[REDACTED]`)
		default:
			result = pat.ReplaceAllString(result, `[REDACTED]`)
		}
	}
	return result
}
