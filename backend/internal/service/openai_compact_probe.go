package service

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// AccountTestModeDefault drives the standard /responses connection test.
	AccountTestModeDefault = "default"
	// AccountTestModeCompact drives the /responses/compact compact-probe test.
	AccountTestModeCompact = "compact"
)

func normalizeAccountTestMode(raw string) string {
	cleaned := strings.ToLower(strings.TrimSpace(raw))
	if cleaned == AccountTestModeCompact {
		return AccountTestModeCompact
	}
	return AccountTestModeDefault
}

func createOpenAICompactProbePayload(modelName string) map[string]any {
	return map[string]any{
		"model":        strings.TrimSpace(modelName),
		"instructions": "You are a helpful coding assistant.",
		"input": []any{
			map[string]any{
				"type":    "message",
				"role":    "user",
				"content": "Respond with OK.",
			},
		},
	}
}

// unsupportedCompactKeywords lists phrases that indicate a compact endpoint is not available.
var unsupportedCompactKeywords = []string{
	"unsupported",
	"not support",
	"does not support",
	"not available",
	"disabled",
}

func shouldMarkOpenAICompactUnsupported(httpStatus int, respBody []byte) bool {
	if httpStatus == http.StatusNotFound || httpStatus == http.StatusMethodNotAllowed || httpStatus == http.StatusNotImplemented {
		return true
	}

	if httpStatus != http.StatusBadRequest && httpStatus != http.StatusForbidden && httpStatus != http.StatusUnprocessableEntity {
		return false
	}

	combined := strings.ToLower(strings.TrimSpace(extractUpstreamErrorMessage(respBody) + " " + string(respBody)))
	if !strings.Contains(combined, "compact") {
		return false
	}
	for _, kw := range unsupportedCompactKeywords {
		if strings.Contains(combined, kw) {
			return true
		}
	}
	return false
}

func buildOpenAICompactProbeExtraUpdates(httpResp *http.Response, respBody []byte, probeFailure error, timestamp time.Time) map[string]any {
	result := map[string]any{
		"openai_compact_checked_at":  timestamp.Format(time.RFC3339),
		"openai_compact_last_status": nil,
	}

	if httpResp != nil {
		result["openai_compact_last_status"] = httpResp.StatusCode
	}

	if probeFailure != nil {
		result["openai_compact_last_error"] = truncateString(sanitizeUpstreamErrorMessage(probeFailure.Error()), 2048)
		return result
	}

	if httpResp == nil {
		result["openai_compact_last_error"] = "compact probe failed"
		return result
	}

	errorText := strings.TrimSpace(extractUpstreamErrorMessage(respBody))
	if errorText == "" && len(respBody) > 0 {
		errorText = strings.TrimSpace(string(respBody))
	}
	if errorText == "" && (httpResp.StatusCode < 200 || httpResp.StatusCode >= 300) {
		errorText = "HTTP " + strconv.Itoa(httpResp.StatusCode)
	}
	errorText = truncateString(sanitizeUpstreamErrorMessage(errorText), 2048)

	statusOK := httpResp.StatusCode >= 200 && httpResp.StatusCode < 300
	if statusOK {
		result["openai_compact_supported"] = true
		result["openai_compact_last_error"] = ""
	} else {
		if shouldMarkOpenAICompactUnsupported(httpResp.StatusCode, respBody) {
			result["openai_compact_supported"] = false
		}
		result["openai_compact_last_error"] = errorText
	}

	return result
}

func mergeExtraUpdates(primary map[string]any, secondary map[string]any) map[string]any {
	if len(primary) == 0 && len(secondary) == 0 {
		return nil
	}
	merged := make(map[string]any, len(primary)+len(secondary))
	for k, v := range primary {
		merged[k] = v
	}
	for k, v := range secondary {
		merged[k] = v
	}
	return merged
}

func compactProbeSessionID(acctID int64) string {
	if acctID <= 0 {
		return "probe_compact"
	}
	return "probe_compact_" + strconv.FormatInt(acctID, 10)
}
