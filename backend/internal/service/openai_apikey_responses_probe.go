package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/telagod/subme/internal/pkg/logger"
	"github.com/telagod/subme/internal/pkg/openai"
	"github.com/telagod/subme/internal/pkg/openai_compat"
)

// openaiResponsesProbeTimeout is the deadline for capability probe requests.
// Probes must fail fast so they do not block account creation workflows.
const openaiResponsesProbeTimeout = 8 * time.Second

// openaiResponsesProbePayload builds a minimal Responses request body used
// solely for endpoint presence detection. Non-streaming mode is used to
// reduce SSE parsing overhead.
//
// Detection logic: only HTTP 404 / 405 means "endpoint absent"; any other
// status (2xx, 4xx like 400/401/422, or 5xx) indicates "endpoint exists".
func openaiResponsesProbePayload(modelID string) []byte {
	if strings.TrimSpace(modelID) == "" {
		modelID = openai.DefaultTestModel
	}
	encoded, _ := json.Marshal(map[string]any{
		"model": modelID,
		"input": []map[string]any{
			{
				"role": "user",
				"content": []map[string]any{
					{"type": "input_text", "text": "hi"},
				},
			},
		},
		"instructions": openai.DefaultInstructions,
		"stream":       false,
	})
	return encoded
}

// ProbeOpenAIAPIKeyResponsesSupport sends a lightweight probe to determine
// whether the upstream supports the /v1/responses endpoint, then persists
// the result to accounts.extra.openai_responses_supported.
//
// This is called after account creation/update for platform=openai,
// type=apikey accounts. All errors are logged but never propagated; probe
// failure must not block the account lifecycle.
func (s *AccountTestService) ProbeOpenAIAPIKeyResponsesSupport(reqCtx context.Context, acctID int64) {
	acct, loadErr := s.accountRepo.GetByID(reqCtx, acctID)
	if loadErr != nil {
		logger.LegacyPrintf("service.openai_probe", "probe_load_account_failed: account_id=%d err=%v", acctID, loadErr)
		return
	}
	if acct.Platform != PlatformOpenAI || acct.Type != AccountTypeAPIKey {
		return
	}

	authKey := acct.GetOpenAIApiKey()
	if authKey == "" {
		logger.LegacyPrintf("service.openai_probe", "probe_skip_no_apikey: account_id=%d", acctID)
		return
	}
	origin := acct.GetOpenAIBaseURL()
	if origin == "" {
		origin = "https://api.openai.com"
	}
	safeURL, urlErr := s.validateUpstreamBaseURL(origin)
	if urlErr != nil {
		logger.LegacyPrintf("service.openai_probe", "probe_invalid_baseurl: account_id=%d base_url=%q err=%v", acctID, origin, urlErr)
		return
	}

	probeEndpoint := buildOpenAIResponsesURL(safeURL)

	probeCtx, cancel := context.WithTimeout(reqCtx, openaiResponsesProbeTimeout)
	defer cancel()

	httpReq, buildErr := http.NewRequestWithContext(probeCtx, http.MethodPost, probeEndpoint, bytes.NewReader(openaiResponsesProbePayload("")))
	if buildErr != nil {
		logger.LegacyPrintf("service.openai_probe", "probe_build_request_failed: account_id=%d err=%v", acctID, buildErr)
		return
	}
	httpReq = httpReq.WithContext(WithHTTPUpstreamProfile(httpReq.Context(), HTTPUpstreamProfileOpenAI))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+authKey)
	httpReq.Header.Set("Accept", "application/json")

	proxyAddr := ""
	if acct.ProxyID != nil && acct.Proxy != nil {
		proxyAddr = acct.Proxy.URL()
	}

	upResp, doErr := s.httpUpstream.DoWithTLS(httpReq, proxyAddr, acct.ID, acct.Concurrency, s.tlsFPProfileService.ResolveTLSProfile(acct))
	if doErr != nil {
		// Network failure: leave the probe result unknown; gateway fallback
		// will handle routing on a per-request basis.
		logger.LegacyPrintf("service.openai_probe", "probe_request_failed: account_id=%d url=%s err=%v", acctID, probeEndpoint, doErr)
		return
	}
	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(upResp.Body, 1<<20))
		_ = upResp.Body.Close()
	}()

	endpointExists := isResponsesEndpointSupportedByStatus(upResp.StatusCode)

	if persistErr := s.accountRepo.UpdateExtra(reqCtx, acctID, map[string]any{
		openai_compat.ExtraKeyResponsesSupported: endpointExists,
	}); persistErr != nil {
		logger.LegacyPrintf("service.openai_probe", "probe_persist_failed: account_id=%d supported=%v err=%v", acctID, endpointExists, persistErr)
		return
	}

	logger.LegacyPrintf("service.openai_probe",
		"probe_done: account_id=%d base_url=%s status=%d supported=%v",
		acctID, safeURL, upResp.StatusCode, endpointExists,
	)
}

// isResponsesEndpointSupportedByStatus infers endpoint presence from the
// HTTP status code. Only 404 and 405 indicate absence; all other statuses
// (including 5xx) are treated as "endpoint exists".
func isResponsesEndpointSupportedByStatus(code int) bool {
	switch code {
	case http.StatusNotFound, http.StatusMethodNotAllowed:
		return false
	}
	return true
}
