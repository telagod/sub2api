package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/telagod/subme/internal/pkg/apicompat"
	"github.com/telagod/subme/internal/pkg/logger"
	"github.com/telagod/subme/internal/util/responseheaders"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.uber.org/zap"
)

// openaiCCRawAllowedHeaders defines the passthrough whitelist for the CC raw
// forwarding path. It intentionally excludes Codex-specific headers
// (originator, session_id, x-codex-turn-state, etc.) that would pollute or
// break third-party OpenAI-compatible upstreams (DeepSeek, Kimi, GLM, etc.).
// Content-Type / Authorization / Accept are set explicitly by the caller.
var openaiCCRawAllowedHeaders = map[string]bool{
	"accept-language": true,
	"user-agent":      true,
}

// forwardAsRawChatCompletions sends the client's Chat Completions request
// directly to the upstream /v1/chat/completions endpoint without any
// CC-to-Responses protocol conversion.
//
// This path is used when the account is platform=openai, type=apikey, and
// the upstream has been probed as NOT supporting /v1/responses (e.g.
// DeepSeek, Kimi, GLM, Qwen, and similar third-party compatible upstreams).
func (s *OpenAIGatewayService) forwardAsRawChatCompletions(
	reqCtx context.Context,
	gc *gin.Context,
	acct *Account,
	rawBody []byte,
	defaultMapped string,
) (*OpenAIForwardResult, error) {
	tStart := time.Now()

	// Step 1: extract routing / billing fields
	srcModel := gjson.GetBytes(rawBody, "model").String()
	if srcModel == "" {
		writeChatCompletionsError(gc, http.StatusBadRequest, "invalid_request_error", "model is required")
		return nil, fmt.Errorf("request body has no model field")
	}
	wantStream := gjson.GetBytes(rawBody, "stream").Bool()

	// Extract reasoning effort and service tier before any body mutation.
	effort := extractOpenAIReasoningEffortFromBody(rawBody, srcModel)
	tier := extractOpenAIServiceTierFromBody(rawBody)

	// Step 2: resolve model names
	billModel := resolveOpenAIForwardModel(acct, srcModel, defaultMapped)
	destModel := normalizeOpenAIModelForUpstream(acct, billModel)

	// Step 3: rewrite model ID (no protocol conversion)
	outBody := rawBody
	if destModel != srcModel {
		outBody = ReplaceModelInBody(rawBody, destModel)
	}

	// Step 4: apply fast policy
	policyBody, policyErr := s.applyOpenAIFastPolicyToBody(reqCtx, acct, destModel, outBody)
	if policyErr != nil {
		var blockedErr *OpenAIFastBlockedError
		if errors.As(policyErr, &blockedErr) {
			MarkOpsClientBusinessLimited(gc, OpsClientBusinessLimitedReasonLocalPolicyDenied)
			writeChatCompletionsError(gc, http.StatusForbidden, "permission_error", blockedErr.Message)
		}
		return nil, policyErr
	}
	outBody = policyBody
	if wantStream {
		var streamErr error
		outBody, streamErr = ensureOpenAIChatStreamUsage(outBody)
		if streamErr != nil {
			return nil, fmt.Errorf("enable stream usage: %w", streamErr)
		}
	}

	logger.L().Debug("openai chat_completions raw: forwarding without protocol conversion",
		zap.Int64("account_id", acct.ID),
		zap.String("original_model", srcModel),
		zap.String("billing_model", billModel),
		zap.String("upstream_model", destModel),
		zap.Bool("stream", wantStream),
	)

	// Step 5: build upstream HTTP request
	authKey := acct.GetOpenAIApiKey()
	if authKey == "" {
		return nil, fmt.Errorf("account %d has no api_key configured", acct.ID)
	}
	origin := acct.GetOpenAIBaseURL()
	if origin == "" {
		origin = "https://api.openai.com"
	}
	safeURL, urlErr := s.validateUpstreamBaseURL(origin)
	if urlErr != nil {
		return nil, fmt.Errorf("invalid base_url: %w", urlErr)
	}
	endpoint := buildOpenAIChatCompletionsURL(safeURL)

	detachedCtx, releaseDetached := decoupleUpstreamContext(reqCtx)
	httpReq, buildErr := http.NewRequestWithContext(detachedCtx, http.MethodPost, endpoint, bytes.NewReader(outBody))
	releaseDetached()
	if buildErr != nil {
		return nil, fmt.Errorf("build upstream request: %w", buildErr)
	}
	httpReq = httpReq.WithContext(WithHTTPUpstreamProfile(httpReq.Context(), HTTPUpstreamProfileOpenAI))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+authKey)
	if wantStream {
		httpReq.Header.Set("Accept", "text/event-stream")
	} else {
		httpReq.Header.Set("Accept", "application/json")
	}

	// Forward whitelisted client headers.
	for hdrName, hdrValues := range gc.Request.Header {
		if openaiCCRawAllowedHeaders[strings.ToLower(hdrName)] {
			for _, hv := range hdrValues {
				httpReq.Header.Add(hdrName, hv)
			}
		}
	}
	if overrideUA := acct.GetOpenAIUserAgent(); overrideUA != "" {
		httpReq.Header.Set("user-agent", overrideUA)
	}

	// Step 6: execute request
	proxyAddr := ""
	if acct.Proxy != nil {
		proxyAddr = acct.Proxy.URL()
	}
	upResp, doErr := s.httpUpstream.Do(httpReq, proxyAddr, acct.ID, acct.Concurrency)
	if doErr != nil {
		sanitized := sanitizeUpstreamErrorMessage(doErr.Error())
		setOpsUpstreamError(gc, 0, sanitized, "")
		appendOpsUpstreamError(gc, OpsUpstreamErrorEvent{
			Platform:           acct.Platform,
			AccountID:          acct.ID,
			AccountName:        acct.Name,
			UpstreamStatusCode: 0,
			Kind:               "request_error",
			Message:            sanitized,
		})
		writeChatCompletionsError(gc, http.StatusBadGateway, "upstream_error", "Upstream request failed")
		return nil, fmt.Errorf("upstream request failed: %s", sanitized)
	}
	defer func() { _ = upResp.Body.Close() }()

	// Step 7: handle error status with failover
	if upResp.StatusCode >= 400 {
		errBody := s.readUpstreamErrorBody(upResp)
		_ = upResp.Body.Close()
		upResp.Body = io.NopCloser(bytes.NewReader(errBody))

		errMsg := sanitizeUpstreamErrorMessage(strings.TrimSpace(extractUpstreamErrorMessage(errBody)))
		if s.shouldFailoverOpenAIUpstreamResponse(upResp.StatusCode, errMsg, errBody) {
			detail := ""
			if s.cfg != nil && s.cfg.Gateway.LogUpstreamErrorBody {
				maxLen := s.cfg.Gateway.LogUpstreamErrorBodyMaxBytes
				if maxLen <= 0 {
					maxLen = 2048
				}
				detail = truncateString(string(errBody), maxLen)
			}
			appendOpsUpstreamError(gc, OpsUpstreamErrorEvent{
				Platform:           acct.Platform,
				AccountID:          acct.ID,
				AccountName:        acct.Name,
				UpstreamStatusCode: upResp.StatusCode,
				UpstreamRequestID:  upResp.Header.Get("x-request-id"),
				Kind:               "failover",
				Message:            errMsg,
				Detail:             detail,
			})
			s.handleOpenAIAccountUpstreamError(reqCtx, acct, upResp.StatusCode, upResp.Header, errBody, destModel)
			return nil, &UpstreamFailoverError{
				StatusCode:             upResp.StatusCode,
				ResponseBody:           errBody,
				RetryableOnSameAccount: acct.IsPoolMode() && (acct.IsPoolModeRetryableStatus(upResp.StatusCode) || isOpenAITransientProcessingError(upResp.StatusCode, errMsg, errBody)),
			}
		}
		return s.handleChatCompletionsErrorResponse(upResp, gc, acct, billModel)
	}

	// Step 8: relay the response
	if wantStream {
		return s.streamRawChatCompletions(gc, upResp, acct, srcModel, billModel, destModel, effort, tier, tStart, len(rawBody))
	}
	return s.bufferRawChatCompletions(gc, upResp, srcModel, billModel, destModel, effort, tier, tStart)
}

// streamRawChatCompletions relays the upstream CC SSE stream to the client
// while extracting usage from the final chunk (per the CC protocol when
// stream_options.include_usage is enabled). The gateway forces include_usage
// on the upstream side and transparently forwards usage downstream.
func (s *OpenAIGatewayService) streamRawChatCompletions(
	gc *gin.Context,
	upResp *http.Response,
	acct *Account,
	srcModel string,
	billModel string,
	destModel string,
	effort *string,
	tier *string,
	tStart time.Time,
	reqBodyLen int,
) (*OpenAIForwardResult, error) {
	reqID := upResp.Header.Get("x-request-id")

	headersSent := false
	emitHeaders := func() {
		if headersSent {
			return
		}
		headersSent = true
		if s.responseHeaderFilter != nil {
			responseheaders.WriteFilteredHeaders(gc.Writer.Header(), upResp.Header, s.responseHeaderFilter)
		}
		gc.Writer.Header().Set("Content-Type", "text/event-stream")
		gc.Writer.Header().Set("Cache-Control", "no-cache")
		gc.Writer.Header().Set("Connection", "keep-alive")
		gc.Writer.Header().Set("X-Accel-Buffering", "no")
		gc.Writer.WriteHeader(http.StatusOK)
	}

	sc := bufio.NewScanner(upResp.Body)
	lineCap := defaultMaxLineSize
	if s.cfg != nil && s.cfg.Gateway.MaxLineSize > 0 {
		lineCap = s.cfg.Gateway.MaxLineSize
	}
	sc.Buffer(make([]byte, 0, 64*1024), lineCap)

	var tokenUsage OpenAIUsage
	var ttft *int
	peerGone := false
	outputStarted := false
	buffered := make([]string, 0, 8)
	refDetector := newOpenAIChatSilentRefusalDetector(reqBodyLen)

	emitLine := func(ln string) {
		if peerGone {
			return
		}
		if !outputStarted && !refDetector.ShouldReleaseClientOutput() {
			buffered = append(buffered, ln)
			return
		}
		if !outputStarted {
			emitHeaders()
			for _, pending := range buffered {
				if _, wErr := gc.Writer.WriteString(pending + "\n"); wErr != nil {
					peerGone = true
					logger.L().Debug("openai chat_completions raw: client disconnected, continuing to drain upstream for billing",
						zap.Error(wErr),
						zap.String("request_id", reqID),
					)
					return
				}
			}
			buffered = buffered[:0]
			outputStarted = true
		}
		if _, wErr := gc.Writer.WriteString(ln + "\n"); wErr != nil {
			peerGone = true
			logger.L().Debug("openai chat_completions raw: client disconnected, continuing to drain upstream for billing",
				zap.Error(wErr),
				zap.String("request_id", reqID),
			)
		}
	}

	for sc.Scan() {
		ln := sc.Text()
		refDetector.ObserveSSELine(ln)
		if data, matched := extractOpenAISSEDataLine(ln); matched {
			stripped := strings.TrimSpace(data)
			if stripped != "[DONE]" {
				onlyUsage := isOpenAIChatUsageOnlyStreamChunk(data)
				if parsed := extractCCStreamUsage(data); parsed != nil {
					tokenUsage = *parsed
				}
				if ttft == nil && !onlyUsage {
					ms := int(time.Since(tStart).Milliseconds())
					ttft = &ms
				}
			}
		}

		emitLine(ln)
		if ln == "" {
			if !peerGone && outputStarted {
				gc.Writer.Flush()
			}
			continue
		}
		if !peerGone && outputStarted {
			gc.Writer.Flush()
		}
	}

	if scanErr := sc.Err(); scanErr != nil {
		if !errors.Is(scanErr, context.Canceled) && !errors.Is(scanErr, context.DeadlineExceeded) {
			logger.L().Warn("openai chat_completions raw: stream read error",
				zap.Error(scanErr),
				zap.String("request_id", reqID),
			)
		}
	} else if !peerGone && !outputStarted {
		if refDetector.IsSilentRefusal() {
			return nil, newOpenAISilentRefusalFailoverError(gc, acct, reqID)
		}
		if len(buffered) > 0 {
			emitHeaders()
			for _, pending := range buffered {
				if _, wErr := gc.Writer.WriteString(pending + "\n"); wErr != nil {
					peerGone = true
					logger.L().Debug("openai chat_completions raw: client disconnected during final flush",
						zap.Error(wErr),
						zap.String("request_id", reqID),
					)
					break
				}
			}
			if !peerGone {
				gc.Writer.Flush()
				outputStarted = true
			}
		}
	}

	return &OpenAIForwardResult{
		RequestID:       reqID,
		Usage:           tokenUsage,
		Model:           srcModel,
		BillingModel:    billModel,
		UpstreamModel:   destModel,
		ReasoningEffort: effort,
		ServiceTier:     tier,
		Stream:          true,
		Duration:        time.Since(tStart),
		FirstTokenMs:    ttft,
	}, nil
}

// ensureOpenAIChatStreamUsage forces include_usage in the stream_options
// so the upstream returns token counts. The flag is also forwarded downstream
// for cascading proxies.
func ensureOpenAIChatStreamUsage(payload []byte) ([]byte, error) {
	out, opErr := sjson.SetBytes(payload, "stream_options.include_usage", true)
	if opErr != nil {
		return payload, opErr
	}
	return out, nil
}

func isOpenAIChatUsageOnlyStreamChunk(data string) bool {
	if strings.TrimSpace(data) == "" {
		return false
	}
	if !gjson.Get(data, "usage").Exists() {
		return false
	}
	choiceArr := gjson.Get(data, "choices")
	return choiceArr.Exists() && choiceArr.IsArray() && len(choiceArr.Array()) == 0
}

// extractCCStreamUsage pulls usage information from a single CC streaming
// chunk. In the CC protocol usage appears only on the final chunk when
// include_usage is active, but the extractor always picks the latest value.
func extractCCStreamUsage(data string) *OpenAIUsage {
	usageNode := gjson.Get(data, "usage")
	if !usageNode.Exists() || !usageNode.IsObject() {
		return nil
	}
	out := OpenAIUsage{
		InputTokens:  int(gjson.Get(data, "usage.prompt_tokens").Int()),
		OutputTokens: int(gjson.Get(data, "usage.completion_tokens").Int()),
	}
	if cachedNode := gjson.Get(data, "usage.prompt_tokens_details.cached_tokens"); cachedNode.Exists() {
		out.CacheReadInputTokens = int(cachedNode.Int())
	}
	return &out
}

// bufferRawChatCompletions relays an upstream non-streaming CC JSON response.
func (s *OpenAIGatewayService) bufferRawChatCompletions(
	gc *gin.Context,
	upResp *http.Response,
	srcModel string,
	billModel string,
	destModel string,
	effort *string,
	tier *string,
	tStart time.Time,
) (*OpenAIForwardResult, error) {
	reqID := upResp.Header.Get("x-request-id")

	bodyBytes, readErr := ReadUpstreamResponseBody(upResp.Body, s.cfg, gc, openAITooLargeError)
	if readErr != nil {
		if !errors.Is(readErr, ErrUpstreamResponseBodyTooLarge) {
			writeChatCompletionsError(gc, http.StatusBadGateway, "api_error", "Failed to read upstream response")
		}
		return nil, fmt.Errorf("read upstream body: %w", readErr)
	}

	var decoded apicompat.ChatCompletionsResponse
	var tokenUsage OpenAIUsage
	if parseErr := json.Unmarshal(bodyBytes, &decoded); parseErr == nil && decoded.Usage != nil {
		tokenUsage = OpenAIUsage{
			InputTokens:  decoded.Usage.PromptTokens,
			OutputTokens: decoded.Usage.CompletionTokens,
		}
		if decoded.Usage.PromptTokensDetails != nil {
			tokenUsage.CacheReadInputTokens = decoded.Usage.PromptTokensDetails.CachedTokens
		}
	}

	if s.responseHeaderFilter != nil {
		responseheaders.WriteFilteredHeaders(gc.Writer.Header(), upResp.Header, s.responseHeaderFilter)
	}
	if contentType := upResp.Header.Get("Content-Type"); contentType != "" {
		gc.Writer.Header().Set("Content-Type", contentType)
	} else {
		gc.Writer.Header().Set("Content-Type", "application/json")
	}
	gc.Writer.WriteHeader(http.StatusOK)
	_, _ = gc.Writer.Write(bodyBytes)

	return &OpenAIForwardResult{
		RequestID:       reqID,
		Usage:           tokenUsage,
		Model:           srcModel,
		BillingModel:    billModel,
		UpstreamModel:   destModel,
		ReasoningEffort: effort,
		ServiceTier:     tier,
		Stream:          false,
		Duration:        time.Since(tStart),
	}, nil
}

// buildOpenAIChatCompletionsURL constructs the upstream CC endpoint URL.
//
//   - If base already ends with /chat/completions, return as-is.
//   - If base ends with a version segment like /v1, append /chat/completions.
//   - Otherwise, append /v1/chat/completions.
func buildOpenAIChatCompletionsURL(base string) string {
	return buildOpenAIEndpointURL(base, "/v1/chat/completions")
}
