package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/telagod/subme/internal/pkg/apicompat"
	"github.com/tidwall/gjson"
)

const (
	openAISilentRefusalMinRequestBodyBytes = 64 * 1024
	openAISilentRefusalErrorCode           = "openai_silent_refusal"
	openAISilentRefusalUpstreamMessage     = "OpenAI upstream returned an empty completion stream with finish_reason=stop and no usage"
	openAISilentRefusalClientMessage       = "Upstream returned an empty completion without usage; no fallback account was available"
)

// openAIChatSilentRefusalDetector monitors an SSE stream for the pattern
// where the upstream emits finish_reason=stop but produces no content,
// tool calls, or usage — indicating a silent refusal.
type openAIChatSilentRefusalDetector struct {
	active          bool
	gotContent      bool
	gotToolCall     bool
	gotFunctionCall bool
	gotUsage        bool
	gotError        bool
	gotReasoning    bool
	gotFinish       bool
	finishReason    string
}

func newOpenAIChatSilentRefusalDetector(bodyLen int) *openAIChatSilentRefusalDetector {
	return &openAIChatSilentRefusalDetector{
		active: bodyLen >= openAISilentRefusalMinRequestBodyBytes,
	}
}

func (det *openAIChatSilentRefusalDetector) Enabled() bool {
	return det != nil && det.active
}

func (det *openAIChatSilentRefusalDetector) ObserveSSELine(ln string) {
	if det == nil || !det.active {
		return
	}
	if evtKind, matched := parseSSEEventLine(ln); matched {
		det.observeEventType(evtKind)
		return
	}
	if data, matched := extractOpenAISSEDataLine(ln); matched {
		det.ObservePayload([]byte(data))
	}
}

func (det *openAIChatSilentRefusalDetector) ObservePayload(data []byte) {
	if det == nil || !det.active {
		return
	}
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("[DONE]")) {
		return
	}
	if !gjson.ValidBytes(data) {
		return
	}

	evtKind := strings.TrimSpace(gjson.GetBytes(data, "type").String())
	det.observeEventType(evtKind)

	if gjson.GetBytes(data, "error").Exists() {
		det.gotError = true
	}
	if node := gjson.GetBytes(data, "usage"); node.Exists() && node.IsObject() {
		det.gotUsage = true
	}
	if node := gjson.GetBytes(data, "response.usage"); node.Exists() && node.IsObject() {
		det.gotUsage = true
	}

	det.observeChatChoicesPayload(data)
	det.observeResponsesPayload(data, evtKind)
}

func (det *openAIChatSilentRefusalDetector) ObserveChatChunk(chk apicompat.ChatCompletionsChunk) {
	if det == nil || !det.active {
		return
	}
	if chk.Usage != nil {
		det.gotUsage = true
	}
	for _, choice := range chk.Choices {
		if choice.FinishReason != nil {
			det.observeFinishReason(*choice.FinishReason)
		}
		delta := choice.Delta
		if delta.Content != nil && *delta.Content != "" {
			det.gotContent = true
		}
		if delta.ReasoningContent != nil {
			det.gotReasoning = true
		}
		if len(delta.ToolCalls) > 0 {
			det.gotToolCall = true
		}
	}
}

// ShouldReleaseClientOutput indicates whether enough signal has been observed
// to safely start writing output to the client.
func (det *openAIChatSilentRefusalDetector) ShouldReleaseClientOutput() bool {
	if det == nil || !det.active {
		return true
	}
	if det.gotContent || det.gotToolCall || det.gotFunctionCall || det.gotUsage || det.gotError || det.gotReasoning {
		return true
	}
	return det.gotFinish && det.finishReason != "" && det.finishReason != "stop"
}

// IsSilentRefusal returns true if the stream completed with finish_reason=stop
// but produced no meaningful output or usage.
func (det *openAIChatSilentRefusalDetector) IsSilentRefusal() bool {
	if det == nil || !det.active {
		return false
	}
	return !det.gotContent &&
		!det.gotToolCall &&
		!det.gotFunctionCall &&
		!det.gotUsage &&
		!det.gotError &&
		!det.gotReasoning &&
		det.gotFinish &&
		det.finishReason == "stop"
}

func (det *openAIChatSilentRefusalDetector) observeEventType(evtKind string) {
	evtKind = strings.TrimSpace(evtKind)
	if evtKind == "" {
		return
	}
	if evtKind == "error" || evtKind == "response.failed" {
		det.gotError = true
	}
	if strings.Contains(evtKind, "reasoning") || strings.Contains(evtKind, "reasoning_summary") {
		det.gotReasoning = true
	}
}

func (det *openAIChatSilentRefusalDetector) observeFinishReason(reason string) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return
	}
	det.gotFinish = true
	det.finishReason = reason
}

func (det *openAIChatSilentRefusalDetector) observeChatChoicesPayload(data []byte) {
	choiceArr := gjson.GetBytes(data, "choices")
	if !choiceArr.Exists() || !choiceArr.IsArray() {
		return
	}
	for _, choice := range choiceArr.Array() {
		if finishNode := choice.Get("finish_reason"); finishNode.Exists() {
			det.observeFinishReason(finishNode.String())
		}
		deltaNode := choice.Get("delta")
		if !deltaNode.Exists() {
			continue
		}
		if textNode := deltaNode.Get("content"); textNode.Exists() && textNode.String() != "" {
			det.gotContent = true
		}
		if deltaNode.Get("tool_calls").Exists() {
			det.gotToolCall = true
		}
		if deltaNode.Get("function_call").Exists() {
			det.gotFunctionCall = true
		}
		if deltaNode.Get("reasoning").Exists() ||
			deltaNode.Get("reasoning_content").Exists() ||
			deltaNode.Get("reasoning_summary").Exists() {
			det.gotReasoning = true
		}
	}
}

func (det *openAIChatSilentRefusalDetector) observeResponsesPayload(data []byte, evtKind string) {
	switch evtKind {
	case "response.output_text.delta":
		if gjson.GetBytes(data, "delta").String() != "" {
			det.gotContent = true
		}
	case "response.output_item.added":
		switch strings.TrimSpace(gjson.GetBytes(data, "item.type").String()) {
		case "function_call":
			det.gotToolCall = true
		case "reasoning":
			det.gotReasoning = true
		}
	case "response.function_call_arguments.delta":
		det.gotToolCall = true
	case "response.reasoning_summary_text.delta", "response.reasoning_summary_text.done":
		det.gotReasoning = true
	case "response.completed", "response.done":
		det.observeFinishReason("stop")
	case "response.incomplete":
		det.observeFinishReason("length")
	case "response.failed":
		det.gotError = true
	}

	if outputArr := gjson.GetBytes(data, "response.output"); outputArr.Exists() && outputArr.IsArray() {
		for _, elem := range outputArr.Array() {
			switch strings.TrimSpace(elem.Get("type").String()) {
			case "function_call":
				det.gotToolCall = true
			case "reasoning":
				det.gotReasoning = true
			case "message":
				det.observeResponseMessageItem(elem)
			}
		}
	}
}

func (det *openAIChatSilentRefusalDetector) observeResponseMessageItem(node gjson.Result) {
	contentArr := node.Get("content")
	if !contentArr.Exists() || !contentArr.IsArray() {
		return
	}
	for _, part := range contentArr.Array() {
		if part.Get("text").String() != "" {
			det.gotContent = true
			return
		}
	}
}

// newOpenAISilentRefusalFailoverError builds a failover error for a silent
// refusal, recording the event in the ops error log.
func newOpenAISilentRefusalFailoverError(gc *gin.Context, acct *Account, upstreamReqID string) *UpstreamFailoverError {
	acctID := int64(0)
	acctName := ""
	plat := PlatformOpenAI
	if acct != nil {
		acctID = acct.ID
		acctName = acct.Name
		plat = acct.Platform
	}

	setOpsUpstreamError(gc, http.StatusBadGateway, openAISilentRefusalUpstreamMessage, "")
	appendOpsUpstreamError(gc, OpsUpstreamErrorEvent{
		Platform:           plat,
		AccountID:          acctID,
		AccountName:        acctName,
		UpstreamStatusCode: http.StatusBadGateway,
		UpstreamRequestID:  upstreamReqID,
		Kind:               "failover",
		Message:            openAISilentRefusalUpstreamMessage,
	})

	hdrs := http.Header{}
	if strings.TrimSpace(upstreamReqID) != "" {
		hdrs.Set("x-request-id", strings.TrimSpace(upstreamReqID))
	}
	return &UpstreamFailoverError{
		StatusCode:      http.StatusBadGateway,
		ResponseBody:    openAISilentRefusalErrorBody(),
		ResponseHeaders: hdrs,
	}
}

func openAISilentRefusalErrorBody() []byte {
	encoded, encErr := json.Marshal(map[string]any{
		"error": map[string]any{
			"type":    "upstream_error",
			"code":    openAISilentRefusalErrorCode,
			"message": openAISilentRefusalUpstreamMessage,
		},
	})
	if encErr != nil {
		return []byte(`{"error":{"type":"upstream_error","code":"openai_silent_refusal","message":"OpenAI upstream returned an empty completion stream with finish_reason=stop and no usage"}}`)
	}
	return encoded
}

// IsOpenAISilentRefusalErrorBody reports whether a failover body was produced
// by the OpenAI silent-refusal detector.
func IsOpenAISilentRefusalErrorBody(body []byte) bool {
	return strings.TrimSpace(gjson.GetBytes(body, "error.code").String()) == openAISilentRefusalErrorCode
}

// OpenAISilentRefusalClientMessage returns the exhausted-failover client message
// for OpenAI silent refusals.
func OpenAISilentRefusalClientMessage() string {
	return openAISilentRefusalClientMessage
}
