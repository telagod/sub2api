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

	"github.com/telagod/subme/internal/pkg/apicompat"
	"github.com/telagod/subme/internal/pkg/logger"
	"github.com/telagod/subme/internal/util/responseheaders"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// forwardResponsesViaRawChatCompletions serves /v1/responses clients through an
// upstream that only supports /v1/chat/completions.
func (s *OpenAIGatewayService) forwardResponsesViaRawChatCompletions(
	reqCtx context.Context,
	gc *gin.Context,
	acct *Account,
	rawBody []byte,
) (*OpenAIForwardResult, error) {
	tStart := time.Now()

	var respReq apicompat.ResponsesRequest
	if parseErr := json.Unmarshal(rawBody, &respReq); parseErr != nil {
		gc.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": "Failed to parse request body",
			},
		})
		return nil, fmt.Errorf("parse responses request: %w", parseErr)
	}
	srcModel := strings.TrimSpace(respReq.Model)
	if srcModel == "" {
		gc.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": "model is required",
			},
		})
		return nil, fmt.Errorf("request body has no model field")
	}

	wantStream := respReq.Stream
	effort := extractOpenAIReasoningEffortFromBody(rawBody, srcModel)
	tier := extractOpenAIServiceTierFromBody(rawBody)

	ccReq, convErr := apicompat.ResponsesToChatCompletionsRequest(&respReq)
	if convErr != nil {
		gc.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": convErr.Error(),
			},
		})
		return nil, fmt.Errorf("convert responses to chat completions: %w", convErr)
	}

	billModel := resolveOpenAIForwardModel(acct, srcModel, "")
	destModel := normalizeOpenAIModelForUpstream(acct, billModel)
	ccReq.Model = destModel
	if wantStream {
		ccReq.StreamOptions = &apicompat.ChatStreamOptions{IncludeUsage: true}
	}

	ccBody, marshalErr := json.Marshal(ccReq)
	if marshalErr != nil {
		return nil, fmt.Errorf("marshal chat completions fallback request: %w", marshalErr)
	}
	ccBody, policyErr := s.applyOpenAIFastPolicyToBody(reqCtx, acct, destModel, ccBody)
	if policyErr != nil {
		var blockedErr *OpenAIFastBlockedError
		if errors.As(policyErr, &blockedErr) {
			writeOpenAIFastPolicyBlockedResponse(gc, blockedErr)
		}
		return nil, policyErr
	}
	if tier == nil {
		tier = extractOpenAIServiceTierFromBody(ccBody)
	}

	logger.L().Debug("openai responses: forwarding via raw chat completions",
		zap.Int64("account_id", acct.ID),
		zap.String("original_model", srcModel),
		zap.String("billing_model", billModel),
		zap.String("upstream_model", destModel),
		zap.Bool("stream", wantStream),
	)

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
	httpReq, buildErr := http.NewRequestWithContext(detachedCtx, http.MethodPost, endpoint, bytes.NewReader(ccBody))
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

	proxyAddr := ""
	if acct.Proxy != nil {
		proxyAddr = acct.Proxy.URL()
	}
	upResp, doErr := s.httpUpstream.Do(httpReq, proxyAddr, acct.ID, acct.Concurrency)
	if doErr != nil {
		// Transport-level failure (proxy/DNS/TCP/TLS). Produce a failover error
		// so the handler can switch to a healthy account.
		return nil, s.handleOpenAIUpstreamTransportError(reqCtx, gc, acct, doErr, false)
	}
	defer func() { _ = upResp.Body.Close() }()

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
		return s.handleErrorResponse(reqCtx, upResp, gc, acct, ccBody, billModel)
	}

	if wantStream {
		return s.streamChatCompletionsAsResponses(gc, upResp, srcModel, billModel, destModel, effort, tier, tStart)
	}
	return s.bufferChatCompletionsAsResponses(gc, upResp, srcModel, billModel, destModel, effort, tier, tStart)
}

func (s *OpenAIGatewayService) bufferChatCompletionsAsResponses(
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
			gc.JSON(http.StatusBadGateway, gin.H{
				"error": gin.H{
					"type":    "api_error",
					"message": "Failed to read upstream response",
				},
			})
		}
		return nil, fmt.Errorf("read upstream body: %w", readErr)
	}

	var decoded apicompat.ChatCompletionsResponse
	if parseErr := json.Unmarshal(bodyBytes, &decoded); parseErr != nil {
		gc.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"type":    "api_error",
				"message": "Failed to parse upstream response",
			},
		})
		return nil, fmt.Errorf("parse chat completions response: %w", parseErr)
	}
	converted := apicompat.ChatCompletionsResponseToResponses(&decoded, srcModel)

	tokenUsage := OpenAIUsage{}
	if parsed, found := extractOpenAIUsageFromJSONBytes(bodyBytes); found {
		tokenUsage = parsed
	}

	if s.responseHeaderFilter != nil {
		responseheaders.WriteFilteredHeaders(gc.Writer.Header(), upResp.Header, s.responseHeaderFilter)
	}
	gc.JSON(http.StatusOK, converted)

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

func (s *OpenAIGatewayService) streamChatCompletionsAsResponses(
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

	converter := apicompat.NewChatCompletionsToResponsesStreamState(srcModel)
	var tokenUsage OpenAIUsage
	var ttft *int
	peerGone := false
	gotDone := false

	emitEvents := func(evts []apicompat.ResponsesStreamEvent) {
		if peerGone || len(evts) == 0 {
			return
		}
		emitHeaders()
		for _, evt := range evts {
			sseBytes, fmtErr := apicompat.ResponsesEventToSSE(evt)
			if fmtErr != nil {
				logger.L().Warn("openai responses chat fallback: failed to marshal stream event",
					zap.Error(fmtErr),
					zap.String("request_id", reqID),
				)
				continue
			}
			if _, wErr := fmt.Fprint(gc.Writer, sseBytes); wErr != nil {
				peerGone = true
				logger.L().Debug("openai responses chat fallback: client disconnected, continuing to drain upstream for billing",
					zap.Error(wErr),
					zap.String("request_id", reqID),
				)
				return
			}
		}
		gc.Writer.Flush()
	}

	sc := bufio.NewScanner(upResp.Body)
	lineCap := defaultMaxLineSize
	if s.cfg != nil && s.cfg.Gateway.MaxLineSize > 0 {
		lineCap = s.cfg.Gateway.MaxLineSize
	}
	sc.Buffer(make([]byte, 0, 64*1024), lineCap)

	for sc.Scan() {
		ln := sc.Text()
		data, matched := extractOpenAISSEDataLine(ln)
		if !matched {
			continue
		}
		data = strings.TrimSpace(data)
		if data == "" {
			continue
		}
		if data == "[DONE]" {
			gotDone = true
			break
		}

		if parsed := extractCCStreamUsage(data); parsed != nil {
			tokenUsage = *parsed
		}

		var chk apicompat.ChatCompletionsChunk
		if unmarshalErr := json.Unmarshal([]byte(data), &chk); unmarshalErr != nil {
			logger.L().Warn("openai responses chat fallback: failed to parse chat stream chunk",
				zap.Error(unmarshalErr),
				zap.String("request_id", reqID),
			)
			continue
		}
		if ttft == nil && !isOpenAIChatUsageOnlyStreamChunk(data) && chatChunkStartsResponsesOutput(&chk) {
			ms := int(time.Since(tStart).Milliseconds())
			ttft = &ms
		}
		emitEvents(apicompat.ChatCompletionsChunkToResponsesEvents(&chk, converter))
	}

	if scanErr := sc.Err(); scanErr != nil {
		if !errors.Is(scanErr, context.Canceled) && !errors.Is(scanErr, context.DeadlineExceeded) {
			logger.L().Warn("openai responses chat fallback: stream read error",
				zap.Error(scanErr),
				zap.String("request_id", reqID),
			)
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
		}, fmt.Errorf("stream usage incomplete: %w", scanErr)
	}

	emitEvents(apicompat.FinalizeChatCompletionsResponsesStream(converter))
	if !peerGone {
		emitHeaders()
		if _, wErr := fmt.Fprint(gc.Writer, "data: [DONE]\n\n"); wErr != nil {
			peerGone = true
		}
		if !peerGone {
			gc.Writer.Flush()
		}
	}
	if !gotDone {
		logger.L().Debug("openai responses chat fallback: upstream stream ended without done sentinel",
			zap.String("request_id", reqID),
		)
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

func chatChunkStartsResponsesOutput(chk *apicompat.ChatCompletionsChunk) bool {
	if chk == nil {
		return false
	}
	for idx := range chk.Choices {
		delta := chk.Choices[idx].Delta
		if delta.Content != nil || delta.ReasoningContent != nil || len(delta.ToolCalls) > 0 {
			return true
		}
	}
	return false
}
