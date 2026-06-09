package service

import (
	"net/url"
	"strings"
)

func buildOpenAIEndpointURL(baseURL string, ep string) string {
	trimmedBase := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	ep = "/" + strings.TrimLeft(strings.TrimSpace(ep), "/")
	withoutVersion := strings.TrimPrefix(ep, "/v1")
	if strings.HasSuffix(trimmedBase, ep) || strings.HasSuffix(trimmedBase, withoutVersion) {
		return trimmedBase
	}
	if openAIBaseURLHasVersionSuffix(trimmedBase) {
		return trimmedBase + withoutVersion
	}
	return trimmedBase + ep
}

func openAIBaseURLHasVersionSuffix(rawURL string) bool {
	cleaned := strings.TrimSpace(rawURL)
	if cleaned == "" {
		return false
	}

	urlPath := ""
	parsed, parseErr := url.Parse(cleaned)
	if parseErr == nil && parsed.Scheme != "" && parsed.Host != "" {
		urlPath = parsed.Path
	} else {
		slashPos := strings.Index(cleaned, "/")
		if slashPos >= 0 {
			urlPath = cleaned[slashPos:]
		}
	}

	urlPath = strings.TrimRight(urlPath, "/")
	if urlPath == "" {
		return false
	}
	lastSep := strings.LastIndex(urlPath, "/")
	tail := urlPath
	if lastSep >= 0 {
		tail = urlPath[lastSep+1:]
	}
	return isOpenAIAPIVersionSegment(tail)
}

func isOpenAIAPIVersionSegment(seg string) bool {
	normalized := strings.ToLower(strings.TrimSpace(seg))
	if len(normalized) < 2 || normalized[0] != 'v' || !isASCIIDigit(normalized[1]) {
		return false
	}

	pos := 1
	for pos < len(normalized) && isASCIIDigit(normalized[pos]) {
		pos++
	}
	if pos == len(normalized) {
		return true
	}
	if normalized[pos] == '.' {
		pos++
		if pos == len(normalized) || !isASCIIDigit(normalized[pos]) {
			return false
		}
		for pos < len(normalized) && isASCIIDigit(normalized[pos]) {
			pos++
		}
		return pos == len(normalized)
	}

	rest := normalized[pos:]
	return strings.HasPrefix(rest, "alpha") ||
		strings.HasPrefix(rest, "beta") ||
		strings.HasPrefix(rest, "preview")
}

func isASCIIDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
