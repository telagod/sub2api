package service

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/telagod/subme/internal/config"
	"github.com/tidwall/gjson"
)

const (
	openAIWSClientReadLimitBytesDefault     int64 = 64 * 1024 * 1024
	openAIWSHTTPBridgeThresholdBytesDefault int64 = 15 * 1024 * 1024
	openAIWSHTTPBridgeErrorBodyLimitBytes         = 64 * 1024
)

// ResolveOpenAIWSClientReadLimitBytes returns the configured read-limit for
// WebSocket client messages, falling back to the 64 MiB default.
func ResolveOpenAIWSClientReadLimitBytes(cfg *config.Config) int64 {
	if cfg == nil || cfg.Gateway.OpenAIWS.ClientReadLimitBytes <= 0 {
		return openAIWSClientReadLimitBytesDefault
	}
	return cfg.Gateway.OpenAIWS.ClientReadLimitBytes
}

func (s *OpenAIGatewayService) openAIWSHTTPBridgeEnabled() bool {
	return s != nil && s.cfg != nil && s.cfg.Gateway.OpenAIWS.HTTPBridgeEnabled
}

func (s *OpenAIGatewayService) openAIWSHTTPBridgeThresholdBytes() int64 {
	if s == nil || s.cfg == nil || s.cfg.Gateway.OpenAIWS.HTTPBridgeThresholdBytes <= 0 {
		return openAIWSHTTPBridgeThresholdBytesDefault
	}
	return s.cfg.Gateway.OpenAIWS.HTTPBridgeThresholdBytes
}

func (s *OpenAIGatewayService) shouldBridgeOpenAIWSHTTP(msgSize int, prevResponseID string) bool {
	if !s.openAIWSHTTPBridgeEnabled() {
		return false
	}
	if strings.TrimSpace(prevResponseID) != "" {
		return false
	}
	limit := s.openAIWSHTTPBridgeThresholdBytes()
	return limit > 0 && int64(msgSize) >= limit
}

func prepareOpenAIWSHTTPBridgeBody(raw []byte) ([]byte, error) {
	var obj map[string]any
	if unmarshalErr := json.Unmarshal(raw, &obj); unmarshalErr != nil {
		return nil, unmarshalErr
	}
	if obj == nil {
		return nil, errors.New("response.create payload must be a JSON object")
	}
	delete(obj, "type")
	delete(obj, "generate")
	delete(obj, "previous_response_id")
	obj["stream"] = true
	return json.Marshal(obj)
}

type openAIWSToolCallReplayCollector struct {
	items []json.RawMessage
	seen  map[string]struct{}
}

func (col *openAIWSToolCallReplayCollector) AddEvent(evtType string, msg []byte) {
	switch strings.TrimSpace(evtType) {
	case "response.output_item.done":
		col.addItem(gjson.GetBytes(msg, "item"))
	case "response.completed", "response.done":
		outputArr := gjson.GetBytes(msg, "response.output")
		if !outputArr.IsArray() {
			return
		}
		for _, elem := range outputArr.Array() {
			col.addItem(elem)
		}
	}
}

func (col *openAIWSToolCallReplayCollector) Items() []json.RawMessage {
	return cloneOpenAIWSRawMessages(col.items)
}

func (col *openAIWSToolCallReplayCollector) addItem(node gjson.Result) {
	if !node.Exists() || node.Type != gjson.JSON {
		return
	}
	rawStr := strings.TrimSpace(node.Raw)
	if rawStr == "" || !strings.HasPrefix(rawStr, "{") {
		return
	}
	if !isCodexToolCallContextItemType(node.Get("type").String()) {
		return
	}
	dedup := strings.TrimSpace(node.Get("id").String())
	if dedup == "" {
		dedup = strings.TrimSpace(node.Get("call_id").String())
	}
	if dedup == "" {
		dedup = rawStr
	}
	if col.seen == nil {
		col.seen = make(map[string]struct{})
	}
	if _, exists := col.seen[dedup]; exists {
		return
	}
	col.seen[dedup] = struct{}{}
	col.items = append(col.items, json.RawMessage(rawStr))
}

func buildOpenAIWSHTTPBridgeErrorEvent(code int, msg string) []byte {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		msg = http.StatusText(code)
	}
	if msg == "" {
		msg = "upstream request failed"
	}
	envelope := map[string]any{
		"type":   "error",
		"status": code,
		"error": map[string]any{
			"type":    "upstream_error",
			"message": msg,
		},
	}
	encoded, encErr := json.Marshal(envelope)
	if encErr != nil {
		return []byte(`{"type":"error","error":{"type":"upstream_error","message":"upstream request failed"}}`)
	}
	return encoded
}

func (s *OpenAIGatewayService) proxyOpenAIWSHTTPBridgeTurn(
	reqCtx context.Context,
	gc *gin.Context,
	acct *Account,
	authToken string,
	rawPayload []byte,
	payloadLen int,
	srcModel string,
	imgBillingModel string,
	imgSizeTier string,
	imgInputSize string,
	turnIdx int,
	sendToClient func([]byte) error,
) (*OpenAIForwardResult, error) {
	if s == nil {
		return nil, errors.New("service is nil")
	}
	if s.httpUpstream == nil {
		return nil, errors.New("openai http upstream is nil")
	}
	if acct == nil {
		return nil, errors.New("account is nil")
	}
	if sendToClient == nil {
		return nil, errors.New("client websocket writer is nil")
	}

	bridgeBody, prepErr := prepareOpenAIWSHTTPBridgeBody(rawPayload)
	if prepErr != nil {
		return nil, fmt.Errorf("prepare http bridge body: %w", prepErr)
	}

	detachedCtx, releaseDetached := decoupleUpstreamContext(reqCtx)
	httpReq, buildErr := s.buildUpstreamRequestOpenAIPassthrough(detachedCtx, gc, acct, bridgeBody, authToken)
	releaseDetached()
	if buildErr != nil {
		return nil, buildErr
	}

	proxyAddr := ""
	if acct.ProxyID != nil && acct.Proxy != nil {
		proxyAddr = acct.Proxy.URL()
	}
	if gc != nil {
		gc.Set("openai_passthrough", true)
		gc.Set("openai_ws_http_bridge", true)
	}

	turnStart := time.Now()
	upResp, doErr := s.httpUpstream.Do(httpReq, proxyAddr, acct.ID, acct.Concurrency)
	if doErr != nil {
		sanitized := sanitizeUpstreamErrorMessage(doErr.Error())
		_ = sendToClient(buildOpenAIWSHTTPBridgeErrorEvent(http.StatusBadGateway, "Upstream request failed"))
		return nil, fmt.Errorf("upstream http bridge request failed: %s", sanitized)
	}
	defer func() { _ = upResp.Body.Close() }()

	if upResp.StatusCode >= 400 {
		errBytes, _ := io.ReadAll(io.LimitReader(upResp.Body, openAIWSHTTPBridgeErrorBodyLimitBytes))
		errMsg := sanitizeUpstreamErrorMessage(strings.TrimSpace(extractUpstreamErrorMessage(errBytes)))
		if errMsg == "" {
			errMsg = http.StatusText(upResp.StatusCode)
		}
		_ = sendToClient(buildOpenAIWSHTTPBridgeErrorEvent(upResp.StatusCode, errMsg))
		return nil, fmt.Errorf("upstream http bridge error: status=%d message=%s", upResp.StatusCode, errMsg)
	}

	respID := ""
	tokenUsage := OpenAIUsage{}
	imgCounter := newOpenAIImageOutputCounter()
	var ttft *int
	isStreaming := openAIWSPayloadBoolFromRaw(bridgeBody, "stream", true)
	totalEvents := 0
	tokenEvents := 0
	terminalEvents := 0
	replay := &openAIWSToolCallReplayCollector{}
	firstEvt := ""
	lastEvt := ""
	gotDone := false
	wroteAny := false
	peerGone := false
	resolvedModel := ""
	doModelReplace := false
	var resolvedModelBytes []byte
	if srcModel != "" {
		resolvedModel = normalizeOpenAIModelForUpstream(acct, acct.GetMappedModel(srcModel))
		doModelReplace = resolvedModel != "" && resolvedModel != srcModel
		if doModelReplace {
			resolvedModelBytes = []byte(resolvedModel)
		}
	}

	buildResult := func() *OpenAIForwardResult {
		imgCount := imgCounter.Count()
		out := &OpenAIForwardResult{
			RequestID:       respID,
			Usage:           tokenUsage,
			Model:           srcModel,
			UpstreamModel:   resolvedModel,
			ServiceTier:     extractOpenAIServiceTierFromBody(bridgeBody),
			ReasoningEffort: extractOpenAIReasoningEffortFromBody(bridgeBody, srcModel),
			Stream:          isStreaming,
			OpenAIWSMode:    true,
			ResponseHeaders: cloneHeader(upResp.Header),
			Duration:        time.Since(turnStart),
			FirstTokenMs:    ttft,
		}
		if replayItems := replay.Items(); len(replayItems) > 0 {
			out.wsReplayInput = replayItems
			out.wsReplayInputExists = true
		}
		if imgCount > 0 {
			out.ImageCount = imgCount
			out.ImageSize = imgSizeTier
			out.ImageInputSize = imgInputSize
			out.ImageOutputSizes = imgCounter.Sizes()
			out.BillingModel = imgBillingModel
		}
		return out
	}

	sc := bufio.NewScanner(upResp.Body)
	lineCap := defaultMaxLineSize
	if s.cfg != nil && s.cfg.Gateway.MaxLineSize > 0 {
		lineCap = s.cfg.Gateway.MaxLineSize
	}
	scanBuf := getSSEScannerBuf64K()
	sc.Buffer(scanBuf[:0], lineCap)
	defer putSSEScannerBuf64K(scanBuf)

	for sc.Scan() {
		ln := sc.Text()
		data, matched := extractOpenAISSEDataLine(ln)
		if !matched {
			continue
		}
		stripped := strings.TrimSpace(data)
		if stripped == "" {
			continue
		}
		if stripped == "[DONE]" {
			gotDone = true
			continue
		}

		evtBytes := []byte(stripped)
		evtKind, evtRespID, _ := parseOpenAIWSEventEnvelope(evtBytes)
		if respID == "" && evtRespID != "" {
			respID = evtRespID
		}
		if evtKind != "" {
			totalEvents++
			if firstEvt == "" {
				firstEvt = evtKind
			}
			lastEvt = evtKind
		}
		if isOpenAIWSTokenEvent(evtKind) {
			tokenEvents++
			if ttft == nil {
				ms := int(time.Since(turnStart).Milliseconds())
				ttft = &ms
			}
		}
		if openAIWSEventShouldParseUsage(evtKind) {
			parseOpenAIWSResponseUsageFromCompletedEvent(evtBytes, &tokenUsage)
		}
		imgCounter.AddSSEData(evtBytes)

		if doModelReplace && len(resolvedModelBytes) > 0 && openAIWSEventMayContainModel(evtKind) && strings.Contains(stripped, resolvedModel) {
			evtBytes = replaceOpenAIWSMessageModel(evtBytes, resolvedModel, srcModel)
		}
		if s.toolCorrector != nil && openAIWSEventMayContainToolCalls(evtKind) && openAIWSMessageLikelyContainsToolCalls(evtBytes) {
			if fixed, changed := s.toolCorrector.CorrectToolCallsInSSEBytes(evtBytes); changed {
				evtBytes = fixed
			}
		}
		replay.AddEvent(evtKind, evtBytes)

		if !peerGone {
			if sendErr := sendToClient(evtBytes); sendErr != nil {
				if isOpenAIWSClientDisconnectError(sendErr) {
					peerGone = true
					closeCode, closeReason := summarizeOpenAIWSReadCloseError(sendErr)
					logOpenAIWSModeInfo(
						"ingress_ws_http_bridge_client_disconnected_drain account_id=%d turn=%d close_status=%s close_reason=%s",
						acct.ID,
						turnIdx,
						closeCode,
						truncateOpenAIWSLogValue(closeReason, openAIWSHeaderValueMaxLen),
					)
				} else {
					return nil, wrapOpenAIWSIngressTurnError(
						"write_client",
						fmt.Errorf("write client websocket event: %w", sendErr),
						wroteAny,
					)
				}
			} else {
				wroteAny = true
			}
		}

		if evtKind == "error" {
			errCodeVal, errTypeVal, errMsgVal := parseOpenAIWSErrorEventFields(evtBytes)
			s.persistOpenAIWSRateLimitSignal(reqCtx, acct, upResp.Header, evtBytes, errCodeVal, errTypeVal, errMsgVal)
			finalMsg := strings.TrimSpace(errMsgVal)
			if finalMsg == "" {
				finalMsg = "upstream error event"
			}
			return buildResult(), errors.New(finalMsg)
		}
		if isOpenAIWSTerminalEvent(evtKind) {
			terminalEvents++
			ttftVal := -1
			if ttft != nil {
				ttftVal = *ttft
			}
			logOpenAIWSModeInfo(
				"ingress_ws_http_bridge_turn_completed account_id=%d turn=%d response_id=%s payload_bytes=%d duration_ms=%d events=%d token_events=%d terminal_events=%d first_event=%s last_event=%s first_token_ms=%d client_disconnected=%v",
				acct.ID,
				turnIdx,
				truncateOpenAIWSLogValue(respID, openAIWSIDValueMaxLen),
				payloadLen,
				time.Since(turnStart).Milliseconds(),
				totalEvents,
				tokenEvents,
				terminalEvents,
				truncateOpenAIWSLogValue(firstEvt, openAIWSLogValueMaxLen),
				truncateOpenAIWSLogValue(lastEvt, openAIWSLogValueMaxLen),
				ttftVal,
				peerGone,
			)
			return buildResult(), nil
		}
	}
	if scanErr := sc.Err(); scanErr != nil {
		return buildResult(), fmt.Errorf("read upstream http bridge stream: %w", scanErr)
	}
	if gotDone && totalEvents > 0 {
		return buildResult(), nil
	}
	return buildResult(), errors.New("upstream http bridge stream ended before terminal event")
}
