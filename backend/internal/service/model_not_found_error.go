package service

import (
	"net/http"
	"strings"
)

var upstreamModelNotFoundKeywords = []string{"model not found", "unknown model", "not found"}

func isUpstreamModelNotFoundError(code int, body []byte) bool {
	if code != http.StatusNotFound {
		return false
	}
	norm := sanitizeModelErrorBody(body)
	if norm == "" {
		return false
	}
	if !strings.Contains(norm, "model") {
		return false
	}
	return matchesModelNotFoundKeyword(norm)
}

func isModelNotFoundError(code int, body []byte) bool {
	if code == http.StatusNotFound {
		return true
	}
	return isUpstreamModelNotFoundError(code, body)
}

func matchesModelNotFoundKeyword(normalized string) bool {
	for _, kw := range upstreamModelNotFoundKeywords {
		if strings.Contains(normalized, kw) {
			return true
		}
	}
	return false
}

func sanitizeModelErrorBody(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	lower := strings.ToLower(string(raw))
	replacer := strings.NewReplacer("_", " ", "-", " ", "\n", " ", "\r", " ", "\t", " ")
	cleaned := replacer.Replace(lower)
	return strings.Join(strings.Fields(cleaned), " ")
}
