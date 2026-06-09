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
	"github.com/telagod/subme/internal/pkg/geminicli"
	"github.com/telagod/subme/internal/pkg/logger"
	"github.com/telagod/subme/internal/util/responseheaders"
)

// ForwardAsChatCompletions serves OpenAI Chat Completions clients through
// Gemini accounts by keeping the client-facing response in Chat Completions
// format while routing the upstream call through Gemini native endpoints.
func (s *GeminiMessagesCompatService) ForwardAsChatCompletions(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	body []byte,
) (*ForwardResult, error) {
	began := time.Now()

	var ccPayload apicompat.ChatCompletionsRequest
	if unmarshalErr := json.Unmarshal(body, &ccPayload); unmarshalErr != nil {
		return nil, s.emitCCError(c, http.StatusBadRequest, "invalid_request_error", "Failed to parse request body")
	}
	if strings.TrimSpace(ccPayload.Model) == "" {
		return nil, s.emitCCError(c, http.StatusBadRequest, "invalid_request_error", "model is required")
	}

	callerModel := ccPayload.Model
	wantStream := ccPayload.Stream
	wantUsageInStream := ccPayload.StreamOptions != nil && ccPayload.StreamOptions.IncludeUsage

	responsesReq, convErr := apicompat.ChatCompletionsToResponses(&ccPayload)
	if convErr != nil {
		return nil, s.emitCCError(c, http.StatusBadRequest, "invalid_request_error", convErr.Error())
	}

	anthropicReq, convErr := apicompat.ResponsesToAnthropicRequest(responsesReq)
	if convErr != nil {
		return nil, s.emitCCError(c, http.StatusBadRequest, "invalid_request_error", convErr.Error())
	}
	anthropicReq.Stream = wantStream

	serialized, marshalErr := json.Marshal(anthropicReq)
	if marshalErr != nil {
		return nil, fmt.Errorf("serialize chat completions compat payload: %w", marshalErr)
	}

	return s.relayAnthropicBodyAsChatCompletions(ctx, c, account, serialized, callerModel, wantStream, wantUsageInStream, began, body)
}

func (s *GeminiMessagesCompatService) relayAnthropicBodyAsChatCompletions(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	anthropicBody []byte,
	callerModel string,
	wantStream bool,
	wantUsageInStream bool,
	began time.Time,
	originalCCBody []byte,
) (*ForwardResult, error) {
	var envelope struct {
		Model  string `json:"model"`
		Stream bool   `json:"stream"`
	}
	if unmarshalErr := json.Unmarshal(anthropicBody, &envelope); unmarshalErr != nil {
		return nil, s.emitCCError(c, http.StatusBadRequest, "invalid_request_error", "Failed to parse request body")
	}
	if strings.TrimSpace(envelope.Model) == "" {
		return nil, s.emitCCError(c, http.StatusBadRequest, "invalid_request_error", "model is required")
	}

	targetModel := envelope.Model
	if account.Type == AccountTypeAPIKey || account.Type == AccountTypeServiceAccount {
		targetModel = account.GetMappedModel(envelope.Model)
	}

	geminiPayload, convErr := convertClaudeMessagesToGeminiGenerateContent(anthropicBody)
	if convErr != nil {
		return nil, s.emitCCError(c, http.StatusBadRequest, "invalid_request_error", convErr.Error())
	}
	geminiPayload = ensureGeminiFunctionCallThoughtSignatures(geminiPayload)

	proxyAddr := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyAddr = account.Proxy.URL()
	}

	upstreamStream := wantStream
	if account.Type == AccountTypeOAuth && !wantStream && strings.TrimSpace(account.GetCredential("project_id")) != "" {
		upstreamStream = true
	}

	makeReq, reqIDHdr := s.buildGeminiCCUpstreamRequestFactory(
		account,
		targetModel,
		geminiPayload,
		wantStream,
		upstreamStream,
	)

	var httpResp *http.Response
	for round := 1; round <= geminiMaxRetries; round++ {
		upReq, idHdr, buildErr := makeReq(ctx)
		if buildErr != nil {
			if errors.Is(buildErr, context.Canceled) || errors.Is(buildErr, context.DeadlineExceeded) {
				return nil, buildErr
			}
			return nil, s.emitCCError(c, http.StatusBadGateway, "upstream_error", buildErr.Error())
		}
		reqIDHdr = idHdr

		var doErr error
		httpResp, doErr = s.httpUpstream.Do(upReq, proxyAddr, account.ID, account.Concurrency)
		if doErr != nil {
			cleanMsg := sanitizeUpstreamErrorMessage(doErr.Error())
			appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
				Platform:           account.Platform,
				AccountID:          account.ID,
				AccountName:        account.Name,
				UpstreamStatusCode: 0,
				Kind:               "request_error",
				Message:            cleanMsg,
			})
			if round < geminiMaxRetries {
				logger.LegacyPrintf("service.gemini_chat_completions", "Gemini account %d: upstream call failed, attempt %d/%d: %v", account.ID, round, geminiMaxRetries, doErr)
				sleepGeminiBackoff(round)
				continue
			}
			setOpsUpstreamError(c, 0, cleanMsg, "")
			return nil, s.emitCCError(c, http.StatusBadGateway, "upstream_error", "Upstream request failed after retries: "+cleanMsg)
		}

		if matched, rebuilt := s.checkErrorPolicyInLoop(ctx, account, httpResp); matched {
			httpResp = rebuilt
			break
		} else {
			httpResp = rebuilt
		}

		if httpResp.StatusCode >= 400 && s.shouldRetryGeminiUpstreamError(account, httpResp.StatusCode) {
			errBody := s.readUpstreamErrorBody(httpResp)
			_ = httpResp.Body.Close()
			if httpResp.StatusCode == http.StatusForbidden && isGeminiInsufficientScope(httpResp.Header, errBody) {
				httpResp = &http.Response{
					StatusCode: httpResp.StatusCode,
					Header:     httpResp.Header.Clone(),
					Body:       io.NopCloser(bytes.NewReader(errBody)),
				}
				break
			}
			if httpResp.StatusCode == http.StatusTooManyRequests {
				s.handleGeminiUpstreamError(ctx, account, httpResp.StatusCode, httpResp.Header, errBody)
			}
			if round < geminiMaxRetries {
				upReqID := httpResp.Header.Get(reqIDHdr)
				if upReqID == "" {
					upReqID = httpResp.Header.Get("x-goog-request-id")
				}
				errText := sanitizeUpstreamErrorMessage(strings.TrimSpace(extractUpstreamErrorMessage(errBody)))
				appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
					Platform:           account.Platform,
					AccountID:          account.ID,
					AccountName:        account.Name,
					UpstreamStatusCode: httpResp.StatusCode,
					UpstreamRequestID:  upReqID,
					Kind:               "retry",
					Message:            errText,
				})
				logger.LegacyPrintf("service.gemini_chat_completions", "Gemini account %d: status %d, attempt %d/%d", account.ID, httpResp.StatusCode, round, geminiMaxRetries)
				sleepGeminiBackoff(round)
				continue
			}
			httpResp = &http.Response{
				StatusCode: httpResp.StatusCode,
				Header:     httpResp.Header.Clone(),
				Body:       io.NopCloser(bytes.NewReader(errBody)),
			}
			break
		}

		break
	}
	defer func() { _ = httpResp.Body.Close() }()

	upReqID := httpResp.Header.Get(reqIDHdr)
	if upReqID == "" {
		upReqID = httpResp.Header.Get("x-goog-request-id")
	}
	if upReqID != "" {
		c.Header("x-request-id", upReqID)
	}

	reasoningLevel := extractCCReasoningEffortFromBody(originalCCBody)

	if httpResp.StatusCode >= 400 {
		errBody := s.readUpstreamErrorBody(httpResp)
		s.handleGeminiUpstreamError(ctx, account, httpResp.StatusCode, httpResp.Header, errBody)
		unwrapped := unwrapIfNeeded(account.Type == AccountTypeOAuth, errBody)

		if s.shouldFailoverGeminiUpstreamError(httpResp.StatusCode) {
			errText := sanitizeUpstreamErrorMessage(strings.TrimSpace(extractUpstreamErrorMessage(unwrapped)))
			appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
				Platform:           account.Platform,
				AccountID:          account.ID,
				AccountName:        account.Name,
				UpstreamStatusCode: httpResp.StatusCode,
				UpstreamRequestID:  upReqID,
				Kind:               "failover",
				Message:            errText,
			})
			return nil, &UpstreamFailoverError{StatusCode: httpResp.StatusCode, ResponseBody: unwrapped}
		}

		return nil, s.mapGeminiCCError(c, account, httpResp.StatusCode, upReqID, unwrapped)
	}

	var consumedUsage *ClaudeUsage
	var ttfMs *int
	if wantStream {
		stResult, stErr := s.handleGeminiCCStream(c, httpResp, began, callerModel, account.Type == AccountTypeOAuth, wantUsageInStream)
		if stErr != nil {
			return nil, stErr
		}
		consumedUsage = stResult.usage
		ttfMs = stResult.firstTokenMs
	} else if upstreamStream {
		gathered, usageObj, gatherErr := collectGeminiSSE(httpResp.Body, account.Type == AccountTypeOAuth)
		if gatherErr != nil {
			return nil, s.emitCCError(c, http.StatusBadGateway, "upstream_error", "Failed to read upstream stream")
		}
		gatheredBytes, _ := json.Marshal(gathered)
		ccResp, usageObj2, convErr := geminiToChatCompletions(gathered, callerModel, gatheredBytes, usageObj)
		if convErr != nil {
			return nil, s.emitCCError(c, http.StatusBadGateway, "upstream_error", "Failed to parse upstream response")
		}
		c.JSON(http.StatusOK, ccResp)
		consumedUsage = usageObj2
	} else {
		syncUsage, syncErr := s.handleGeminiCCSync(c, httpResp, callerModel, account.Type == AccountTypeOAuth)
		if syncErr != nil {
			return nil, syncErr
		}
		consumedUsage = syncUsage
	}

	if consumedUsage == nil {
		consumedUsage = &ClaudeUsage{}
	}

	imgCount := 0
	imgInputSize := s.extractImageInputSize(anthropicBody)
	imgTier := deriveImageSizeTier(imgInputSize)
	if isImageGenerationModel(callerModel) {
		imgCount = 1
	}

	return &ForwardResult{
		RequestID:        upReqID,
		Usage:            *consumedUsage,
		Model:            callerModel,
		UpstreamModel:    targetModel,
		Stream:           wantStream,
		Duration:         time.Since(began),
		FirstTokenMs:     ttfMs,
		ReasoningEffort:  reasoningLevel,
		ImageCount:       imgCount,
		ImageSize:        imgTier,
		ImageInputSize:   imgInputSize,
		ClientDisconnect: false,
	}, nil
}

func (s *GeminiMessagesCompatService) buildGeminiCCUpstreamRequestFactory(
	account *Account,
	targetModel string,
	geminiPayload []byte,
	wantStream bool,
	upstreamStream bool,
) (func(context.Context) (*http.Request, string, error), string) {
	switch account.Type {
	case AccountTypeAPIKey:
		return func(ctx context.Context) (*http.Request, string, error) {
			apiKey := account.GetCredential("api_key")
			if strings.TrimSpace(apiKey) == "" {
				return nil, "", errors.New("gemini api_key is not configured")
			}

			baseAddr := account.GetGeminiBaseURL(geminicli.AIStudioBaseURL)
			normalizedBase, valErr := s.validateUpstreamBaseURL(baseAddr)
			if valErr != nil {
				return nil, "", valErr
			}

			verb := "generateContent"
			if wantStream {
				verb = "streamGenerateContent"
			}
			endpoint := fmt.Sprintf("%s/v1beta/models/%s:%s", strings.TrimRight(normalizedBase, "/"), targetModel, verb)
			if wantStream {
				endpoint += "?alt=sse"
			}

			normalizedPayload := normalizeGeminiRequestForAIStudio(geminiPayload)
			httpReq, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(normalizedPayload))
			if reqErr != nil {
				return nil, "", reqErr
			}
			httpReq.Header.Set("Content-Type", "application/json")
			httpReq.Header.Set("x-goog-api-key", apiKey)
			return httpReq, "x-request-id", nil
		}, "x-request-id"

	case AccountTypeOAuth:
		return func(ctx context.Context) (*http.Request, string, error) {
			if s.tokenProvider == nil {
				return nil, "", errors.New("gemini token provider is not configured")
			}
			tok, tokErr := s.tokenProvider.GetAccessToken(ctx, account)
			if tokErr != nil {
				return nil, "", tokErr
			}

			projectID := strings.TrimSpace(account.GetCredential("project_id"))
			verb := "generateContent"
			if upstreamStream {
				verb = "streamGenerateContent"
			}

			if projectID != "" {
				baseAddr, valErr := s.validateUpstreamBaseURL(geminicli.GeminiCliBaseURL)
				if valErr != nil {
					return nil, "", valErr
				}
				endpoint := fmt.Sprintf("%s/v1internal:%s", strings.TrimRight(baseAddr, "/"), verb)
				if upstreamStream {
					endpoint += "?alt=sse"
				}

				var innerPayload any
				if unmarshalErr := json.Unmarshal(geminiPayload, &innerPayload); unmarshalErr != nil {
					return nil, "", fmt.Errorf("unable to parse gemini request: %w", unmarshalErr)
				}
				wrappedBytes, _ := json.Marshal(map[string]any{
					"model":   targetModel,
					"project": projectID,
					"request": innerPayload,
				})

				httpReq, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(wrappedBytes))
				if reqErr != nil {
					return nil, "", reqErr
				}
				httpReq.Header.Set("Content-Type", "application/json")
				httpReq.Header.Set("Authorization", "Bearer "+tok)
				httpReq.Header.Set("User-Agent", geminicli.GeminiCLIUserAgent)
				return httpReq, "x-request-id", nil
			}

			baseAddr := account.GetGeminiBaseURL(geminicli.AIStudioBaseURL)
			normalizedBase, valErr := s.validateUpstreamBaseURL(baseAddr)
			if valErr != nil {
				return nil, "", valErr
			}

			endpoint := fmt.Sprintf("%s/v1beta/models/%s:%s", strings.TrimRight(normalizedBase, "/"), targetModel, verb)
			if upstreamStream {
				endpoint += "?alt=sse"
			}

			normalizedPayload := normalizeGeminiRequestForAIStudio(geminiPayload)
			httpReq, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(normalizedPayload))
			if reqErr != nil {
				return nil, "", reqErr
			}
			httpReq.Header.Set("Content-Type", "application/json")
			httpReq.Header.Set("Authorization", "Bearer "+tok)
			return httpReq, "x-request-id", nil
		}, "x-request-id"

	case AccountTypeServiceAccount:
		return func(ctx context.Context) (*http.Request, string, error) {
			if s.tokenProvider == nil {
				return nil, "", errors.New("gemini token provider is not configured")
			}
			tok, tokErr := s.tokenProvider.GetAccessToken(ctx, account)
			if tokErr != nil {
				return nil, "", tokErr
			}

			verb := "generateContent"
			if wantStream {
				verb = "streamGenerateContent"
			}
			endpoint, urlErr := buildVertexGeminiURL(account.VertexProjectID(), account.VertexLocation(targetModel), targetModel, verb, wantStream)
			if urlErr != nil {
				return nil, "", urlErr
			}

			normalizedPayload := normalizeGeminiRequestForAIStudio(geminiPayload)
			httpReq, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(normalizedPayload))
			if reqErr != nil {
				return nil, "", reqErr
			}
			httpReq.Header.Set("Content-Type", "application/json")
			httpReq.Header.Set("Authorization", "Bearer "+tok)
			return httpReq, "x-request-id", nil
		}, "x-request-id"

	default:
		return func(context.Context) (*http.Request, string, error) {
			return nil, "", fmt.Errorf("account type %s is not supported", account.Type)
		}, "x-request-id"
	}
}

func (s *GeminiMessagesCompatService) handleGeminiCCSync(
	c *gin.Context,
	resp *http.Response,
	callerModel string,
	isOAuth bool,
) (*ClaudeUsage, error) {
	rawBody, readErr := ReadUpstreamResponseBody(resp.Body, s.cfg, c, openAITooLargeError)
	if readErr != nil {
		return nil, readErr
	}
	if isOAuth {
		if inner, uwErr := unwrapGeminiResponse(rawBody); uwErr == nil {
			rawBody = inner
		}
	}

	var geminiMap map[string]any
	if unmarshalErr := json.Unmarshal(rawBody, &geminiMap); unmarshalErr != nil {
		return nil, s.emitCCError(c, http.StatusBadGateway, "upstream_error", "Failed to parse upstream response")
	}

	ccResp, usageObj, convErr := geminiToChatCompletions(geminiMap, callerModel, rawBody, nil)
	if convErr != nil {
		return nil, s.emitCCError(c, http.StatusBadGateway, "upstream_error", "Failed to convert upstream response")
	}

	responseheaders.WriteFilteredHeaders(c.Writer.Header(), resp.Header, s.responseHeaderFilter)
	c.JSON(http.StatusOK, ccResp)
	return usageObj, nil
}

func geminiToChatCompletions(
	geminiMap map[string]any,
	callerModel string,
	rawPayload []byte,
	usageOverride *ClaudeUsage,
) (*apicompat.ChatCompletionsResponse, *ClaudeUsage, error) {
	claudeMap, usageObj := convertGeminiToClaudeMessage(geminiMap, callerModel, rawPayload)
	if usageOverride != nil && (usageOverride.InputTokens > 0 || usageOverride.OutputTokens > 0 || usageOverride.CacheReadInputTokens > 0) {
		usageObj = usageOverride
		if uMap, ok := claudeMap["usage"].(map[string]any); ok {
			uMap["input_tokens"] = usageObj.InputTokens
			uMap["output_tokens"] = usageObj.OutputTokens
			uMap["cache_read_input_tokens"] = usageObj.CacheReadInputTokens
		}
	}

	claudeBytes, marshalErr := json.Marshal(claudeMap)
	if marshalErr != nil {
		return nil, nil, marshalErr
	}
	var anthropicResp apicompat.AnthropicResponse
	if unmarshalErr := json.Unmarshal(claudeBytes, &anthropicResp); unmarshalErr != nil {
		return nil, nil, unmarshalErr
	}
	responsesResp := apicompat.AnthropicToResponsesResponse(&anthropicResp)
	return apicompat.ResponsesToChatCompletions(responsesResp, callerModel), usageObj, nil
}

func (s *GeminiMessagesCompatService) handleGeminiCCStream(
	c *gin.Context,
	resp *http.Response,
	began time.Time,
	callerModel string,
	isOAuth bool,
	wantUsageInStream bool,
) (*geminiStreamResult, error) {
	if s.responseHeaderFilter != nil {
		responseheaders.WriteFilteredHeaders(c.Writer.Header(), resp.Header, s.responseHeaderFilter)
	}
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)

	fl, flOK := c.Writer.(http.Flusher)
	if !flOK {
		return nil, errors.New("response writer does not support streaming")
	}

	anthState := apicompat.NewAnthropicEventToResponsesState()
	anthState.Model = callerModel
	ccState := apicompat.NewResponsesEventToChatState()
	ccState.Model = callerModel
	ccState.IncludeUsage = wantUsageInStream

	var accUsage ClaudeUsage
	var ttfMs *int
	isFirstChunk := true

	emitCCChunk := func(chunk apicompat.ChatCompletionsChunk) bool {
		encoded, encErr := apicompat.ChatChunkToSSE(chunk)
		if encErr != nil {
			return false
		}
		if _, wErr := io.WriteString(c.Writer, encoded); wErr != nil {
			return true
		}
		return false
	}

	pushAnthropicEvent := func(evt *apicompat.AnthropicStreamEvent) bool {
		for _, resEvt := range apicompat.AnthropicEventToResponsesEvents(evt, anthState) {
			for _, chunk := range apicompat.ResponsesEventToChatChunks(&resEvt, ccState) {
				if gone := emitCCChunk(chunk); gone {
					return true
				}
			}
		}
		fl.Flush()
		return false
	}

	msgID := "msg_" + randomHex(12)
	if pushAnthropicEvent(&apicompat.AnthropicStreamEvent{
		Type: "message_start",
		Message: &apicompat.AnthropicResponse{
			ID:      msgID,
			Type:    "message",
			Role:    "assistant",
			Model:   callerModel,
			Content: []apicompat.AnthropicContentBlock{},
			Usage:   apicompat.AnthropicUsage{},
		},
	}) {
		return &geminiStreamResult{usage: &accUsage, firstTokenMs: ttfMs}, nil
	}

	stopReason := ""
	sawToolUse := false
	blockSeq := 0
	activeBlockIdx := -1
	activeBlockKind := ""
	cumulativeText := ""
	activeToolIdx := -1
	activeToolName := ""
	cumulativeToolJSON := ""

	closeBlock := func() bool {
		if activeBlockIdx < 0 {
			return false
		}
		gone := pushAnthropicEvent(&apicompat.AnthropicStreamEvent{Type: "content_block_stop"})
		activeBlockIdx = -1
		activeBlockKind = ""
		return gone
	}
	closeTool := func() bool {
		if activeToolIdx < 0 {
			return false
		}
		gone := pushAnthropicEvent(&apicompat.AnthropicStreamEvent{Type: "content_block_stop"})
		activeToolIdx = -1
		activeToolName = ""
		cumulativeToolJSON = ""
		return gone
	}

	scanner := bufio.NewReader(resp.Body)
	for {
		rawLine, readErr := scanner.ReadString('\n')
		if len(rawLine) > 0 {
			stripped := strings.TrimRight(rawLine, "\r\n")
			if strings.HasPrefix(stripped, "data:") {
				data := strings.TrimSpace(strings.TrimPrefix(stripped, "data:"))
				if data != "" && data != "[DONE]" {
					chunk := []byte(data)
					if isOAuth {
						if inner, uwErr := unwrapGeminiResponse(chunk); uwErr == nil {
							chunk = inner
						}
					}

					var geminiChunk map[string]any
					if unmarshalErr := json.Unmarshal(chunk, &geminiChunk); unmarshalErr == nil {
						if isFirstChunk {
							isFirstChunk = false
							elapsed := int(time.Since(began).Milliseconds())
							ttfMs = &elapsed
						}
						if fr := extractGeminiFinishReason(geminiChunk); fr != "" {
							stopReason = fr
						}
						if u := extractGeminiUsage(chunk); u != nil {
							accUsage = *u
						}

						for _, part := range extractGeminiParts(geminiChunk) {
							if textVal, ok := part["text"].(string); ok && textVal != "" {
								if activeToolIdx >= 0 {
									if closeTool() {
										return &geminiStreamResult{usage: &accUsage, firstTokenMs: ttfMs}, nil
									}
								}
								delta, newCum := computeGeminiTextDelta(cumulativeText, textVal)
								cumulativeText = newCum
								if delta == "" {
									continue
								}
								if activeBlockKind != "text" {
									if closeBlock() {
										return &geminiStreamResult{usage: &accUsage, firstTokenMs: ttfMs}, nil
									}
									idx := blockSeq
									blockSeq++
									activeBlockIdx = idx
									activeBlockKind = "text"
									if pushAnthropicEvent(&apicompat.AnthropicStreamEvent{
										Type:  "content_block_start",
										Index: &idx,
										ContentBlock: &apicompat.AnthropicContentBlock{
											Type: "text",
											Text: "",
										},
									}) {
										return &geminiStreamResult{usage: &accUsage, firstTokenMs: ttfMs}, nil
									}
								}
								if pushAnthropicEvent(&apicompat.AnthropicStreamEvent{
									Type: "content_block_delta",
									Delta: &apicompat.AnthropicDelta{
										Type: "text_delta",
										Text: delta,
									},
								}) {
									return &geminiStreamResult{usage: &accUsage, firstTokenMs: ttfMs}, nil
								}
								continue
							}

							if fcMap, ok := part["functionCall"].(map[string]any); ok && fcMap != nil {
								fnName, _ := fcMap["name"].(string)
								if strings.TrimSpace(fnName) == "" {
									fnName = "tool"
								}
								if closeBlock() {
									return &geminiStreamResult{usage: &accUsage, firstTokenMs: ttfMs}, nil
								}
								if activeToolIdx >= 0 && activeToolName != fnName {
									if closeTool() {
										return &geminiStreamResult{usage: &accUsage, firstTokenMs: ttfMs}, nil
									}
								}
								if activeToolIdx < 0 {
									idx := blockSeq
									blockSeq++
									activeToolIdx = idx
									activeToolName = fnName
									sawToolUse = true
									if pushAnthropicEvent(&apicompat.AnthropicStreamEvent{
										Type:  "content_block_start",
										Index: &idx,
										ContentBlock: &apicompat.AnthropicContentBlock{
											Type:  "tool_use",
											ID:    "toolu_" + randomHex(8),
											Name:  fnName,
											Input: json.RawMessage(`{}`),
										},
									}) {
										return &geminiStreamResult{usage: &accUsage, firstTokenMs: ttfMs}, nil
									}
								}

								argsText := "{}"
								switch av := fcMap["args"].(type) {
								case nil:
								case string:
									if strings.TrimSpace(av) != "" {
										argsText = av
									}
								default:
									if encoded, encErr := json.Marshal(av); encErr == nil && len(encoded) > 0 {
										argsText = string(encoded)
									}
								}
								argsDelta, newCum := computeGeminiTextDelta(cumulativeToolJSON, argsText)
								cumulativeToolJSON = newCum
								if argsDelta != "" {
									if pushAnthropicEvent(&apicompat.AnthropicStreamEvent{
										Type: "content_block_delta",
										Delta: &apicompat.AnthropicDelta{
											Type:        "input_json_delta",
											PartialJSON: argsDelta,
										},
									}) {
										return &geminiStreamResult{usage: &accUsage, firstTokenMs: ttfMs}, nil
									}
								}
							}
						}
					}
				}
			}
		}

		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("stream read error: %w", readErr)
		}
	}

	if closeBlock() {
		return &geminiStreamResult{usage: &accUsage, firstTokenMs: ttfMs}, nil
	}
	if closeTool() {
		return &geminiStreamResult{usage: &accUsage, firstTokenMs: ttfMs}, nil
	}

	claudeStop := mapGeminiFinishReasonToClaudeStopReason(stopReason)
	if sawToolUse {
		claudeStop = "tool_use"
	}
	anthState.InputTokens = accUsage.InputTokens
	anthState.CacheReadInputTokens = accUsage.CacheReadInputTokens
	if pushAnthropicEvent(&apicompat.AnthropicStreamEvent{
		Type: "message_delta",
		Delta: &apicompat.AnthropicDelta{
			Type:       "message_delta",
			StopReason: claudeStop,
		},
		Usage: &apicompat.AnthropicUsage{
			InputTokens:          accUsage.InputTokens,
			OutputTokens:         accUsage.OutputTokens,
			CacheReadInputTokens: accUsage.CacheReadInputTokens,
		},
	}) {
		return &geminiStreamResult{usage: &accUsage, firstTokenMs: ttfMs}, nil
	}
	if pushAnthropicEvent(&apicompat.AnthropicStreamEvent{Type: "message_stop"}) {
		return &geminiStreamResult{usage: &accUsage, firstTokenMs: ttfMs}, nil
	}

	for _, resEvt := range apicompat.FinalizeAnthropicResponsesStream(anthState) {
		for _, chunk := range apicompat.ResponsesEventToChatChunks(&resEvt, ccState) {
			if gone := emitCCChunk(chunk); gone {
				return &geminiStreamResult{usage: &accUsage, firstTokenMs: ttfMs}, nil
			}
		}
	}
	for _, chunk := range apicompat.FinalizeResponsesChatStream(ccState) {
		if gone := emitCCChunk(chunk); gone {
			return &geminiStreamResult{usage: &accUsage, firstTokenMs: ttfMs}, nil
		}
	}

	_, _ = io.WriteString(c.Writer, "data: [DONE]\n\n")
	fl.Flush()

	return &geminiStreamResult{usage: &accUsage, firstTokenMs: ttfMs}, nil
}

func (s *GeminiMessagesCompatService) mapGeminiCCError(
	c *gin.Context,
	account *Account,
	upstreamStatus int,
	upReqID string,
	body []byte,
) error {
	errText := sanitizeUpstreamErrorMessage(strings.TrimSpace(extractUpstreamErrorMessage(body)))
	setOpsUpstreamError(c, upstreamStatus, errText, "")
	if account != nil {
		appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
			Platform:           account.Platform,
			AccountID:          account.ID,
			AccountName:        account.Name,
			UpstreamStatusCode: upstreamStatus,
			UpstreamRequestID:  upReqID,
			Kind:               "http_error",
			Message:            errText,
		})
	}

	if status, eType, eMsg, matched := applyErrorPassthroughRule(
		c,
		PlatformGemini,
		upstreamStatus,
		body,
		http.StatusBadGateway,
		"upstream_error",
		"Upstream request failed",
	); matched {
		return s.emitCCError(c, status, eType, eMsg)
	}

	httpCode := http.StatusBadGateway
	eType := "upstream_error"
	eMsg := "Upstream request failed"
	if mapped := mapGeminiErrorBodyToClaudeError(body); mapped != nil {
		if mapped.Type != "" {
			eType = mapped.Type
		}
		if mapped.Message != "" {
			eMsg = mapped.Message
		}
		if mapped.StatusCode > 0 {
			httpCode = mapped.StatusCode
		}
	}

	switch upstreamStatus {
	case http.StatusBadRequest:
		if httpCode == http.StatusBadGateway {
			httpCode = http.StatusBadRequest
		}
		if eType == "upstream_error" {
			eType = "invalid_request_error"
		}
		if eMsg == "Upstream request failed" {
			eMsg = "Invalid request"
		}
	case http.StatusNotFound:
		httpCode = http.StatusNotFound
		if eType == "upstream_error" {
			eType = "not_found_error"
		}
		if eMsg == "Upstream request failed" {
			eMsg = "Resource not found"
		}
	case http.StatusTooManyRequests:
		httpCode = http.StatusTooManyRequests
		if eType == "upstream_error" {
			eType = "rate_limit_error"
		}
		if eMsg == "Upstream request failed" {
			eMsg = "Upstream rate limit exceeded, please retry later"
		}
	case 529:
		httpCode = http.StatusServiceUnavailable
		if eType == "upstream_error" {
			eType = "overloaded_error"
		}
		if eMsg == "Upstream request failed" {
			eMsg = "Upstream service overloaded, please retry later"
		}
	}

	if errText != "" && eMsg == "Upstream request failed" {
		eMsg = errText
	}
	return s.emitCCError(c, httpCode, eType, eMsg)
}

func (s *GeminiMessagesCompatService) emitCCError(c *gin.Context, status int, errType, message string) error {
	c.JSON(status, gin.H{
		"error": gin.H{
			"type":    errType,
			"message": message,
		},
	})
	return fmt.Errorf("%s", message)
}
