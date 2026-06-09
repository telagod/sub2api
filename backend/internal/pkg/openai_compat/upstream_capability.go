// Package openai_compat provides capability detection utilities for OpenAI-compatible
// upstream providers.
//
// Background: sub2api's OpenAI APIKey accounts connect to various third-party
// OpenAI-compatible upstreams (DeepSeek, Kimi, GLM, Qwen, etc.) via base_url.
// Most of these only support /v1/chat/completions and lack a /v1/responses endpoint.
// Historical gateway code unconditionally converted CC to Responses and forwarded
// to /v1/responses, causing 404 errors on incompatible upstreams.
//
// This package uses account-level probe markers, set once at account creation
// or modification by internal/service/openai_apikey_responses_probe.go, to
// determine upstream capabilities at routing time.
//
// Design rationale:
//   - No static host allowlist is maintained to avoid code changes when new
//     providers appear.
//   - When the probe marker is absent, the default is true (use Responses),
//     preserving identical behavior to the pre-refactor codebase for existing
//     accounts ("current state is evidence").
package openai_compat

// AccountResponsesSupport describes the effective support state of an account's
// upstream for the OpenAI Responses API.
//
// Only applicable to platform=openai + type=apikey accounts.
type AccountResponsesSupport int

const (
	// ResponsesSupportUnknown indicates the account has not been probed yet
	// (extra field is missing). The routing layer should default to Responses
	// to preserve pre-refactor behavior.
	ResponsesSupportUnknown AccountResponsesSupport = iota

	// ResponsesSupportYes indicates the probe confirmed /v1/responses is available.
	ResponsesSupportYes

	// ResponsesSupportNo indicates the probe confirmed /v1/responses is NOT
	// available; the direct /v1/chat/completions path should be used instead.
	ResponsesSupportNo
)

// ResponsesSupportMode represents the account-level Responses API routing override.
type ResponsesSupportMode string

const (
	// ResponsesSupportModeAuto follows the automatic probe result.
	ResponsesSupportModeAuto ResponsesSupportMode = "auto"

	// ResponsesSupportModeForceResponses forces use of /v1/responses.
	ResponsesSupportModeForceResponses ResponsesSupportMode = "force_responses"

	// ResponsesSupportModeForceChatCompletions forces use of /v1/chat/completions.
	ResponsesSupportModeForceChatCompletions ResponsesSupportMode = "force_chat_completions"
)

// ExtraKeyResponsesMode is the JSON key in accounts.extra that stores the manual
// routing override. Values: auto, force_responses, force_chat_completions.
const ExtraKeyResponsesMode = "openai_responses_mode"

// ExtraKeyResponsesSupported is the JSON key in accounts.extra that stores the
// automatic probe result. Type: bool (true=supported, false=unsupported, absent=unprobed).
const ExtraKeyResponsesSupported = "openai_responses_supported"

// NormalizeResponsesSupportMode canonicalizes the routing override mode string.
// Unrecognized or empty values fall back to auto to preserve existing behavior.
func NormalizeResponsesSupportMode(mode string) ResponsesSupportMode {
	candidate := ResponsesSupportMode(mode)
	switch candidate {
	case ResponsesSupportModeForceResponses:
		return ResponsesSupportModeForceResponses
	case ResponsesSupportModeForceChatCompletions:
		return ResponsesSupportModeForceChatCompletions
	default:
		return ResponsesSupportModeAuto
	}
}

// ResolveResponsesSupport reads the manual override and probe marker from the
// account's extra map to determine the effective support state.
//
// When the marker is missing or has an unexpected type, ResponsesSupportUnknown
// is returned. Callers should treat unknown as "use Responses" to preserve
// pre-refactor behavior (see ShouldUseResponsesAPI).
func ResolveResponsesSupport(extra map[string]any) AccountResponsesSupport {
	if extra == nil {
		return ResponsesSupportUnknown
	}
	// Check for manual override first.
	if rawMode, found := extra[ExtraKeyResponsesMode].(string); found {
		normalized := NormalizeResponsesSupportMode(rawMode)
		switch normalized {
		case ResponsesSupportModeForceResponses:
			return ResponsesSupportYes
		case ResponsesSupportModeForceChatCompletions:
			return ResponsesSupportNo
		}
	}
	// Fall back to automatic probe result.
	rawVal, exists := extra[ExtraKeyResponsesSupported]
	if !exists {
		return ResponsesSupportUnknown
	}
	probeResult, isBool := rawVal.(bool)
	if !isBool {
		return ResponsesSupportUnknown
	}
	if probeResult {
		return ResponsesSupportYes
	}
	return ResponsesSupportNo
}

// ShouldUseResponsesAPI determines whether an OpenAI APIKey account's inbound
// /v1/chat/completions request should be routed through the CC-to-Responses
// conversion path.
//
// Returns true in two cases:
//  1. The probe confirmed Responses support.
//  2. The account has not been probed (marker absent) -- preserves legacy behavior.
//
// Returns false only when the probe explicitly confirmed lack of support.
func ShouldUseResponsesAPI(extra map[string]any) bool {
	return ResolveResponsesSupport(extra) != ResponsesSupportNo
}
