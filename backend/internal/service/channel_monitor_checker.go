package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

// Shared HTTP clients for monitor checks; avoids rebuilding transport per check.
// Custom transport forces IP re-validation at dial time to prevent DNS rebinding.
var monitorHTTPClient = newSSRFSafeHTTPClient(monitorRequestTimeout)

// Separate client for endpoint origin HEAD pings with shorter timeout.
var monitorPingHTTPClient = newSSRFSafeHTTPClient(monitorPingTimeout)

// newSSRFSafeHTTPClient returns an http.Client using safeDialContext.
// Only used by the monitor subsystem for outbound requests to public endpoints.
func newSSRFSafeHTTPClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		DialContext:           safeDialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          16,
		IdleConnTimeout:       monitorIdleConnTimeout,
		TLSHandshakeTimeout:   monitorTLSHandshakeTimeout,
		ResponseHeaderTimeout: monitorResponseHeaderTimeout,
	}
	return &http.Client{Timeout: timeout, Transport: transport}
}

// CheckOptions carries optional per-check customization.
// All fields are optional (zero value = use default behavior).
type CheckOptions struct {
	// APIMode only applies to OpenAI provider; empty string equals chat_completions.
	APIMode string
	// ExtraHeaders are user-defined HTTP headers merged onto adapter defaults (user takes priority).
	ExtraHeaders map[string]string
	// BodyOverrideMode: off | merge | replace
	BodyOverrideMode string
	// BodyOverride is shallow-merged in merge mode (deny-listed keys silently dropped)
	// or used as-is in replace mode.
	BodyOverride map[string]any
}

// runCheckForModel performs a complete check for a single (provider, model) pair.
// Never returns error: all failures are encoded into CheckResult.Status=error/failed.
//
// opts carries template/monitor snapshot customization; nil equals "off + no extra headers".
func runCheckForModel(reqCtx context.Context, prov, endpoint, apiKey, modelName string, opts *CheckOptions) *CheckResult {
	cr := &CheckResult{
		Model:     modelName,
		Status:    MonitorStatusError,
		CheckedAt: time.Now(),
	}

	chal := generateChallenge()
	overrideMode := bodyOverrideMode(opts)

	t0 := time.Now()
	respContent, rawResp, httpStatus, callErr := callProvider(reqCtx, prov, endpoint, apiKey, modelName, chal.Prompt, opts)
	elapsed := time.Since(t0)
	elapsedMs := int(elapsed / time.Millisecond)
	cr.LatencyMs = &elapsedMs

	if callErr != nil {
		cr.Status = MonitorStatusError
		cr.Message = truncateMessage(sanitizeErrorMessage(callErr.Error()))
		return cr
	}
	if httpStatus < 200 || httpStatus >= 300 {
		// Error path: use rawResp rather than extracted text (gjson textPath extraction
		// on error responses usually yields empty string, losing the real upstream error).
		cr.Status = MonitorStatusError
		snippet := truncateForErrorBody(rawResp)
		cr.Message = truncateMessage(sanitizeErrorMessage(fmt.Sprintf("upstream HTTP %d: %s", httpStatus, snippet)))
		return cr
	}

	// Replace mode: skip challenge validation (user body is static, challenge can't be embedded).
	// Use "HTTP 2xx + non-empty extracted text" as operational indicator.
	// Empty text degrades to failed (upstream returned 200 but no actual content).
	if overrideMode == MonitorBodyOverrideModeReplace {
		if strings.TrimSpace(respContent) == "" {
			cr.Status = MonitorStatusFailed
			cr.Message = truncateMessage("replace-mode: upstream returned 2xx with empty text")
			return cr
		}
		return finalizeOperationalOrDegraded(cr, elapsed, elapsedMs)
	}

	if !validateChallenge(respContent, chal.Expected) {
		cr.Status = MonitorStatusFailed
		cr.Message = truncateMessage(sanitizeErrorMessage(fmt.Sprintf("challenge mismatch (expected %s, got %q)", chal.Expected, respContent)))
		return cr
	}

	return finalizeOperationalOrDegraded(cr, elapsed, elapsedMs)
}

// finalizeOperationalOrDegraded handles the final operational/degraded determination.
func finalizeOperationalOrDegraded(cr *CheckResult, elapsed time.Duration, elapsedMs int) *CheckResult {
	if elapsed >= monitorDegradedThreshold {
		cr.Status = MonitorStatusDegraded
		cr.Message = truncateMessage(fmt.Sprintf("slow response: %dms", elapsedMs))
		return cr
	}
	cr.Status = MonitorStatusOperational
	return cr
}

// bodyOverrideMode normalizes opts.BodyOverrideMode; nil opts or empty string → off.
func bodyOverrideMode(opts *CheckOptions) string {
	if opts == nil || opts.BodyOverrideMode == "" {
		return MonitorBodyOverrideModeOff
	}
	return opts.BodyOverrideMode
}

// pingEndpointOrigin issues a HEAD request to the endpoint origin, returning latency in ms.
// Returns nil on failure (does not affect main status determination).
func pingEndpointOrigin(reqCtx context.Context, endpoint string) *int {
	origin, extractErr := extractOrigin(endpoint)
	if extractErr != nil || origin == "" {
		return nil
	}
	headReq, reqErr := http.NewRequestWithContext(reqCtx, http.MethodHead, origin, nil)
	if reqErr != nil {
		return nil
	}
	t0 := time.Now()
	headResp, doErr := monitorPingHTTPClient.Do(headReq)
	if doErr != nil {
		return nil
	}
	defer func() { _ = headResp.Body.Close() }()
	_, _ = io.Copy(io.Discard, io.LimitReader(headResp.Body, monitorPingDiscardMaxBytes))
	latencyMs := int(time.Since(t0) / time.Millisecond)
	return &latencyMs
}

// providerAdapter describes the four things needed for a provider in challenge checks:
//   - build the request path (may include model placeholder)
//   - serialize the request body
//   - construct auth headers
//   - gjson path for extracting response text
//
// Adding a new provider only requires adding an entry to providerAdapters.
type providerAdapter struct {
	buildPath    func(modelName string) string
	buildBody    func(modelName, prompt string) ([]byte, error)
	buildHeaders func(apiKey string) map[string]string
	textPath     string
}

// providerAdapters maps all supported providers. Keys are MonitorProvider* string constants.
//
//nolint:gochecknoglobals // Read-only static adapter table, immutable after init.
var providerAdapters = map[string]providerAdapter{
	MonitorProviderOpenAI: providerOpenAIChatAdapter,
	MonitorProviderAnthropic: {
		buildPath: func(string) string { return providerAnthropicPath },
		buildBody: func(modelName, prompt string) ([]byte, error) {
			return json.Marshal(map[string]any{
				"model":      modelName,
				"messages":   []map[string]string{{"role": "user", "content": prompt}},
				"max_tokens": monitorChallengeMaxTokens,
			})
		},
		buildHeaders: func(apiKey string) map[string]string {
			return map[string]string{
				"x-api-key":         apiKey,
				"anthropic-version": monitorAnthropicAPIVersion,
			}
		},
		textPath: "content.0.text",
	},
	MonitorProviderGemini: {
		buildPath: func(modelName string) string { return fmt.Sprintf(providerGeminiPathTemplate, modelName) },
		buildBody: func(_, prompt string) ([]byte, error) {
			return json.Marshal(map[string]any{
				"contents": []map[string]any{
					{"parts": []map[string]any{{"text": prompt}}},
				},
				"generationConfig": map[string]any{"maxOutputTokens": monitorChallengeMaxTokens},
			})
		},
		// Uses x-goog-api-key header instead of ?key= query to prevent *url.Error from
		// leaking the key into error logs.
		buildHeaders: func(apiKey string) map[string]string {
			return map[string]string{"x-goog-api-key": apiKey}
		},
		textPath: "candidates.0.content.parts.0.text",
	},
}

//nolint:gochecknoglobals // Read-only static adapter table, immutable after init.
var providerOpenAIChatAdapter = providerAdapter{
	buildPath: func(string) string { return providerOpenAIPath },
	buildBody: func(modelName, prompt string) ([]byte, error) {
		return json.Marshal(map[string]any{
			"model":      modelName,
			"messages":   []map[string]string{{"role": "user", "content": prompt}},
			"max_tokens": monitorChallengeMaxTokens,
			"stream":     false,
		})
	},
	buildHeaders: func(apiKey string) map[string]string {
		return map[string]string{"Authorization": "Bearer " + apiKey}
	},
	textPath: "choices.0.message.content",
}

//nolint:gochecknoglobals // Read-only static adapter table, immutable after init.
var providerOpenAIResponsesAdapter = providerAdapter{
	buildPath: func(string) string { return providerOpenAIResponsesPath },
	buildBody: func(modelName, prompt string) ([]byte, error) {
		return json.Marshal(map[string]any{
			"model":             modelName,
			"instructions":      "You are a channel health-check endpoint. Answer the arithmetic challenge exactly and briefly.",
			"input":             prompt,
			"max_output_tokens": monitorChallengeMaxTokens,
			"stream":            false,
		})
	},
	buildHeaders: func(apiKey string) map[string]string {
		return map[string]string{"Authorization": "Bearer " + apiKey}
	},
	textPath: "output.0.content.0.text",
}

// providerAdapterFor selects the concrete adapter based on provider + api_mode.
func providerAdapterFor(prov, apiMode string) (providerAdapter, string, bool) {
	if prov == MonitorProviderOpenAI && defaultAPIMode(apiMode) == MonitorAPIModeResponses {
		return providerOpenAIResponsesAdapter, MonitorAPIModeResponses, true
	}
	adapter, found := providerAdapters[prov]
	return adapter, MonitorAPIModeChatCompletions, found
}

// isSupportedProvider checks whether the provider string exists in the adapter table.
// Used by validateProvider to avoid maintaining a separate switch statement.
func isSupportedProvider(prov string) bool {
	_, found := providerAdapters[prov]
	return found
}

// callProvider dispatches to the appropriate provider adapter.
// opts carries user-defined headers/body overrides (may be nil).
//
// Returns:
//   - extractedText: text extracted via textPath, meaningful only on 2xx; usually empty on error
//   - rawBody: full response body string (capped at monitorResponseMaxBytes) for error path
//   - status: HTTP status code
//   - err: network/serialization errors
func callProvider(reqCtx context.Context, prov, endpoint, apiKey, modelName, prompt string, opts *CheckOptions) (extractedText, rawBody string, status int, opErr error) {
	requestedMode := checkAPIMode(opts)
	if validationErr := validateAPIMode(prov, requestedMode); validationErr != nil {
		return "", "", 0, validationErr
	}
	adapter, resolvedMode, found := providerAdapterFor(prov, requestedMode)
	if !found {
		return "", "", 0, fmt.Errorf("unsupported provider %q", prov)
	}
	payload, bodyErr := buildRequestBody(adapter, prov, resolvedMode, modelName, prompt, opts)
	if bodyErr != nil {
		return "", "", 0, bodyErr
	}
	hdrs := mergeHeaders(adapter.buildHeaders(apiKey), opts)
	fullURL := joinURL(endpoint, adapter.buildPath(modelName))
	respBytes, httpCode, postErr := postRawJSON(reqCtx, fullURL, payload, hdrs)
	if postErr != nil {
		return "", "", httpCode, postErr
	}
	if prov == MonitorProviderOpenAI && resolvedMode == MonitorAPIModeResponses {
		return extractOpenAIResponsesText(respBytes), string(respBytes), httpCode, nil
	}
	return gjson.GetBytes(respBytes, adapter.textPath).String(), string(respBytes), httpCode, nil
}

// extractOpenAIResponsesText aggregates the final assistant text from Responses API output.
// The output array order is model-determined (reasoning/tool-call items may precede message),
// so we cannot assume text is always at output.0.content.0.text.
func extractOpenAIResponsesText(respBytes []byte) string {
	if topLevel := gjson.GetBytes(respBytes, "output_text").String(); strings.TrimSpace(topLevel) != "" {
		return topLevel
	}

	var collected []string
	outputArr := gjson.GetBytes(respBytes, "output")
	if outputArr.IsArray() {
		outputArr.ForEach(func(_, outputItem gjson.Result) bool {
			itemType := outputItem.Get("type").String()
			if itemType != "" && itemType != "message" {
				return true
			}

			contentArr := outputItem.Get("content")
			if !contentArr.IsArray() {
				return true
			}

			contentArr.ForEach(func(_, block gjson.Result) bool {
				blockType := block.Get("type").String()
				if blockType != "" && blockType != "output_text" {
					return true
				}
				if txt := block.Get("text").String(); strings.TrimSpace(txt) != "" {
					collected = append(collected, txt)
				}
				return true
			})
			return true
		})
	}

	if len(collected) > 0 {
		return strings.Join(collected, "")
	}
	return gjson.GetBytes(respBytes, providerOpenAIResponsesAdapter.textPath).String()
}

// mergeHeaders combines user-defined extra headers onto adapter default headers.
// User values override defaults; deny-listed (hop-by-hop / client-managed) keys are silently dropped.
func mergeHeaders(baseHdrs map[string]string, opts *CheckOptions) map[string]string {
	if opts == nil || len(opts.ExtraHeaders) == 0 {
		return baseHdrs
	}
	merged := make(map[string]string, len(baseHdrs)+len(opts.ExtraHeaders))
	for hdrKey, hdrVal := range baseHdrs {
		merged[hdrKey] = hdrVal
	}
	for hdrKey, hdrVal := range opts.ExtraHeaders {
		if IsForbiddenHeaderName(hdrKey) {
			continue
		}
		merged[hdrKey] = hdrVal
	}
	return merged
}

// buildRequestBody constructs the request body based on body_override_mode:
//   - off:     adapter default body
//   - merge:   adapter default body shallow-merged with BodyOverride (deny-listed keys dropped)
//   - replace: BodyOverride marshaled directly as the complete body
//
// Returns valid JSON bytes ready for postRawJSON.
func buildRequestBody(adapter providerAdapter, prov, apiMode, modelName, prompt string, opts *CheckOptions) ([]byte, error) {
	overrideMode := bodyOverrideMode(opts)

	if overrideMode == MonitorBodyOverrideModeReplace {
		if opts == nil || len(opts.BodyOverride) == 0 {
			return nil, fmt.Errorf("replace mode: body_override is empty")
		}
		if replaceErr := validateReplaceRequestBody(prov, apiMode, opts.BodyOverride); replaceErr != nil {
			return nil, replaceErr
		}
		serialized, marshalErr := json.Marshal(opts.BodyOverride)
		if marshalErr != nil {
			return nil, fmt.Errorf("marshal body_override (replace): %w", marshalErr)
		}
		return serialized, nil
	}

	defaultPayload, defaultErr := adapter.buildBody(modelName, prompt)
	if defaultErr != nil {
		return nil, fmt.Errorf("marshal default body: %w", defaultErr)
	}
	if overrideMode != MonitorBodyOverrideModeMerge || opts == nil || len(opts.BodyOverride) == 0 {
		return defaultPayload, nil
	}

	var baseMap map[string]any
	if unmarshalErr := json.Unmarshal(defaultPayload, &baseMap); unmarshalErr != nil {
		return nil, fmt.Errorf("unmarshal default body for merge: %w", unmarshalErr)
	}
	denySet := bodyMergeKeyDenyList[bodyMergeDenyKey(prov, apiMode)]
	for overrideKey, overrideVal := range opts.BodyOverride {
		if denySet[overrideKey] {
			continue
		}
		baseMap[overrideKey] = overrideVal
	}
	mergedPayload, mergeErr := json.Marshal(baseMap)
	if mergeErr != nil {
		return nil, fmt.Errorf("marshal merged body: %w", mergeErr)
	}
	return mergedPayload, nil
}

// bodyMergeKeyDenyList prevents users from overriding provider-specific critical fields in merge mode.
// Users who need to control these fields should use replace mode (which skips challenge validation).
//
//nolint:gochecknoglobals // Static lookup table, immutable after init.
var bodyMergeKeyDenyList = map[string]map[string]bool{
	MonitorProviderOpenAI + ":" + MonitorAPIModeChatCompletions: {"model": true, "messages": true, "stream": true},
	MonitorProviderOpenAI + ":" + MonitorAPIModeResponses:       {"model": true, "instructions": true, "input": true, "stream": true},
	MonitorProviderAnthropic:                                    {"model": true, "messages": true},
	MonitorProviderGemini:                                       {"contents": true},
}

func checkAPIMode(opts *CheckOptions) string {
	if opts == nil {
		return MonitorAPIModeChatCompletions
	}
	return defaultAPIMode(opts.APIMode)
}

func bodyMergeDenyKey(prov, apiMode string) string {
	if prov == MonitorProviderOpenAI {
		return prov + ":" + defaultAPIMode(apiMode)
	}
	return prov
}

func validateReplaceRequestBody(prov, apiMode string, body map[string]any) error {
	if prov != MonitorProviderOpenAI {
		return nil
	}
	switch defaultAPIMode(apiMode) {
	case MonitorAPIModeResponses:
		if strings.TrimSpace(stringFromAny(body["instructions"])) == "" || !hasNonEmptyBodyValue(body["input"]) {
			return fmt.Errorf("replace mode responses body: instructions and input are required")
		}
	case MonitorAPIModeChatCompletions:
		if !hasNonEmptyBodyValue(body["messages"]) {
			return fmt.Errorf("replace mode chat_completions body: messages are required")
		}
	}
	return nil
}

func stringFromAny(val any) string {
	str, _ := val.(string)
	return str
}

func hasNonEmptyBodyValue(val any) bool {
	switch typed := val.(type) {
	case nil:
		return false
	case string:
		return strings.TrimSpace(typed) != ""
	case []any:
		return len(typed) > 0
	case []map[string]any:
		return len(typed) > 0
	case []map[string]string:
		return len(typed) > 0
	default:
		return true
	}
}

// postRawJSON sends a POST with pre-serialized JSON bytes, caps response body size,
// and returns response bytes, HTTP status, and error.
// Adapters marshal their own body for precise field ordering/typing, so this accepts []byte.
func postRawJSON(reqCtx context.Context, targetURL string, payload []byte, hdrs map[string]string) ([]byte, int, error) {
	httpReq, buildErr := http.NewRequestWithContext(reqCtx, http.MethodPost, targetURL, bytes.NewReader(payload))
	if buildErr != nil {
		return nil, 0, fmt.Errorf("build request: %w", buildErr)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	for hdrKey, hdrVal := range hdrs {
		httpReq.Header.Set(hdrKey, hdrVal)
	}

	httpResp, doErr := monitorHTTPClient.Do(httpReq)
	if doErr != nil {
		return nil, 0, fmt.Errorf("do request: %w", doErr)
	}
	defer func() { _ = httpResp.Body.Close() }()

	respBytes, readErr := io.ReadAll(io.LimitReader(httpResp.Body, monitorResponseMaxBytes))
	if readErr != nil {
		return nil, httpResp.StatusCode, fmt.Errorf("read body: %w", readErr)
	}
	return respBytes, httpResp.StatusCode, nil
}

// joinURL combines a base origin and a path into a full URL.
// Tolerates trailing slash on base; path must have leading slash.
func joinURL(base, pathSegment string) string {
	trimmedBase := strings.TrimRight(base, "/")
	if !strings.HasPrefix(pathSegment, "/") {
		pathSegment = "/" + pathSegment
	}
	return trimmedBase + pathSegment
}

// extractOrigin extracts scheme://host[:port] from an endpoint URL.
func extractOrigin(endpoint string) (string, error) {
	parsed, parseErr := url.Parse(endpoint)
	if parseErr != nil {
		return "", parseErr
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("endpoint missing scheme or host")
	}
	return parsed.Scheme + "://" + parsed.Host, nil
}

// Regex matching URL query params that may leak credentials.
var monitorSensitiveQueryParamRegex = regexp.MustCompile(`(?i)([?&](?:key|api[_-]?key|access[_-]?token|token|authorization|x-api-key)=)[^&\s"']+`)

// Patterns matching common provider API key literals.
// Order matters: sk-ant- must precede sk- to avoid being consumed by the generic pattern.
var monitorAPIKeyPatterns = []struct {
	pattern *regexp.Regexp
	replace string
}{
	// Anthropic (prefixed, must match first): sk-ant-xxxxxxx
	{regexp.MustCompile(`sk-ant-[A-Za-z0-9_-]{20,}`), "sk-ant-***REDACTED***"},
	// OpenAI / Anthropic generic sk-: sk-xxxxxxx
	{regexp.MustCompile(`sk-[A-Za-z0-9-]{20,}`), "sk-***REDACTED***"},
	// Gemini / Google API Key: fixed prefix + 35 chars
	{regexp.MustCompile(`AIza[A-Za-z0-9_-]{35}`), "AIza***REDACTED***"},
	// JWT three-segment (often follows Bearer): eyJxxx.eyJxxx.signature
	{regexp.MustCompile(`eyJ[A-Za-z0-9_-]{8,}\.eyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}`), "eyJ***REDACTED.JWT***"},
}

// sanitizeErrorMessage scrubs potential API key leaks from error/response text.
// Handles two sources:
//  1. URL query params (?key=, ?api_key=, etc.) — Go *url.Error may include full URL
//  2. Key fragments in upstream HTTP body (sk-*, AIza*, JWT, etc.)
func sanitizeErrorMessage(rawMsg string) string {
	if rawMsg == "" {
		return rawMsg
	}
	scrubbed := monitorSensitiveQueryParamRegex.ReplaceAllString(rawMsg, `${1}REDACTED`)
	for idx := 0; idx < len(monitorAPIKeyPatterns); idx++ {
		scrubbed = monitorAPIKeyPatterns[idx].pattern.ReplaceAllString(scrubbed, monitorAPIKeyPatterns[idx].replace)
	}
	return scrubbed
}

// truncateMessage caps the message at monitorMessageMaxBytes to prevent DB column overflow.
func truncateMessage(rawMsg string) string {
	if len(rawMsg) <= monitorMessageMaxBytes {
		return rawMsg
	}
	const suffix = "...(truncated)"
	limit := monitorMessageMaxBytes - len(suffix)
	if limit < 0 {
		limit = 0
	}
	return rawMsg[:limit] + suffix
}

// truncateForErrorBody compresses the upstream error body to monitorErrorBodySnippetMaxBytes,
// collapsing consecutive whitespace (upstream HTML error pages often contain excessive indentation).
// Final overall truncation is handled by truncateMessage downstream.
func truncateForErrorBody(rawBody string) string {
	compacted := strings.Join(strings.Fields(rawBody), " ")
	if len(compacted) <= monitorErrorBodySnippetMaxBytes {
		return compacted
	}
	const suffix = "...(body truncated)"
	limit := monitorErrorBodySnippetMaxBytes - len(suffix)
	if limit < 0 {
		limit = 0
	}
	return compacted[:limit] + suffix
}
