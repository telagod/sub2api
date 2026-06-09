package apicompat

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ResponsesToChatCompletionsRequest converts a Responses API request into a
// Chat Completions request for upstreams that only implement
// /v1/chat/completions.
func ResponsesToChatCompletionsRequest(req *ResponsesRequest) (*ChatCompletionsRequest, error) {
	if req == nil {
		return nil, fmt.Errorf("responses request is nil")
	}

	chatMsgs, convertErr := transformInputToChatMessages(req.Instructions, req.Input)
	if convertErr != nil {
		return nil, convertErr
	}

	result := &ChatCompletionsRequest{
		Model:               req.Model,
		Messages:            chatMsgs,
		MaxCompletionTokens: req.MaxOutputTokens,
		Temperature:         req.Temperature,
		TopP:                req.TopP,
		Stream:              req.Stream,
		ServiceTier:         req.ServiceTier,
	}
	if req.Reasoning != nil {
		result.ReasoningEffort = req.Reasoning.Effort
	}
	if len(req.Tools) > 0 {
		result.Tools = convertResponsesTools(req.Tools)
	}
	if len(req.ToolChoice) > 0 {
		result.ToolChoice = convertResponsesToolChoice(req.ToolChoice)
	}

	return result, nil
}

// transformInputToChatMessages converts a Responses request's instructions +
// input[] into Chat Completions messages. It is a three-stage pipeline:
//
//	parse   — instructions become a system message; input[] is split into items
//	build   — assembleMessagesFromItems walks items, attaching reasoning to the
//	          assistant message that produced a tool call, merging parallel tool
//	          calls into one assistant message, and skipping item types that have
//	          no Chat equivalent
//	normalize — reorderToolCallMessages enforces the invariants DeepSeek requires
//
// The build + normalize split keeps every protocol rule in one place rather than
// scattered across per-item cases, and makes unknown future codex item types
// fail safe instead of leaking into the upstream request.
func transformInputToChatMessages(instructions string, rawInput json.RawMessage) ([]ChatMessage, error) {
	var chatMsgs []ChatMessage
	if strings.TrimSpace(instructions) != "" {
		encoded, _ := json.Marshal(instructions)
		chatMsgs = append(chatMsgs, ChatMessage{Role: "system", Content: encoded})
	}

	rawInput = trimJSONWhitespace(rawInput)
	if len(rawInput) == 0 || string(rawInput) == "null" {
		return chatMsgs, nil
	}

	// Bare string input is a single user turn.
	var plainText string
	if unmarshalErr := json.Unmarshal(rawInput, &plainText); unmarshalErr == nil {
		encoded, _ := json.Marshal(plainText)
		chatMsgs = append(chatMsgs, ChatMessage{Role: "user", Content: encoded})
		return chatMsgs, nil
	}

	var itemArray []json.RawMessage
	if unmarshalErr := json.Unmarshal(rawInput, &itemArray); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to parse responses input array: %w", unmarshalErr)
	}

	assembled, buildErr := assembleMessagesFromItems(chatMsgs, itemArray)
	if buildErr != nil {
		return nil, buildErr
	}
	return reorderToolCallMessages(assembled), nil
}

// assembleMessagesFromItems walks the Responses input items and appends the
// corresponding Chat messages.
func assembleMessagesFromItems(chatMsgs []ChatMessage, itemArray []json.RawMessage) ([]ChatMessage, error) {
	// bufferedReasoning holds the reasoning text from a reasoning item until the
	// assistant message it belongs to is emitted. DeepSeek's thinking mode
	// requires the reasoning_content that produced a tool call to be passed back
	// on that assistant message; dropping it yields a 400. It only survives
	// across an assistant message (so a following tool call in the same turn
	// still receives it); any other role ends the thinking span.
	var bufferedReasoning string

	for _, rawItem := range itemArray {
		rawItem = trimJSONWhitespace(rawItem)
		if len(rawItem) == 0 || string(rawItem) == "null" {
			continue
		}

		var fields map[string]json.RawMessage
		if parseErr := json.Unmarshal(rawItem, &fields); parseErr != nil {
			var plainStr string
			if strErr := json.Unmarshal(rawItem, &plainStr); strErr == nil {
				encoded, _ := json.Marshal(plainStr)
				chatMsgs = append(chatMsgs, ChatMessage{Role: "user", Content: encoded})
				bufferedReasoning = ""
				continue
			}
			return nil, fmt.Errorf("failed to parse responses input item: %w", parseErr)
		}

		msgRole := mapResponsesRole(extractRawString(fields["role"]))
		itemKind := extractRawString(fields["type"])
		switch itemKind {
		case "reasoning":
			if reasoningText := gatherReasoningText(fields); reasoningText != "" {
				bufferedReasoning = reasoningText
			}
			continue
		case "function_call":
			fnArgs := extractRawString(fields["arguments"])
			if strings.TrimSpace(fnArgs) == "" {
				fnArgs = "{}"
			}
			call := ChatToolCall{
				ID:   extractRawString(fields["call_id"]),
				Type: "function",
				Function: ChatFunctionCall{
					Name:      extractRawString(fields["name"]),
					Arguments: fnArgs,
				},
			}
			// Parallel tool calls arrive as consecutive function_call items and
			// must share one assistant message; the matching tool replies then
			// follow it. Merge into the immediately preceding assistant message.
			if msgCount := len(chatMsgs); msgCount > 0 && chatMsgs[msgCount-1].Role == "assistant" {
				chatMsgs[msgCount-1].ToolCalls = append(chatMsgs[msgCount-1].ToolCalls, call)
				if chatMsgs[msgCount-1].ReasoningContent == "" {
					chatMsgs[msgCount-1].ReasoningContent = bufferedReasoning
				}
			} else {
				chatMsgs = append(chatMsgs, ChatMessage{
					Role:             "assistant",
					ToolCalls:        []ChatToolCall{call},
					ReasoningContent: bufferedReasoning,
				})
			}
			bufferedReasoning = ""
			continue
		case "function_call_output":
			encoded, _ := json.Marshal(extractRawString(fields["output"]))
			chatMsgs = append(chatMsgs, ChatMessage{
				Role:       "tool",
				ToolCallID: extractRawString(fields["call_id"]),
				Content:    encoded,
			})
			bufferedReasoning = ""
			continue
		case "input_text", "text":
			encoded, _ := json.Marshal(extractRawString(fields["text"]))
			chatMsgs = append(chatMsgs, ChatMessage{Role: "user", Content: encoded})
			bufferedReasoning = ""
			continue
		case "input_image":
			imgContent, imgErr := convertSingleResponsesPart(itemKind, fields)
			if imgErr != nil {
				return nil, imgErr
			}
			chatMsgs = append(chatMsgs, ChatMessage{Role: "user", Content: imgContent})
			bufferedReasoning = ""
			continue
		}

		// Only genuine message items become chat messages. Codex emits other
		// Responses item types with no Chat equivalent (web_search_call,
		// local_shell_call, custom tool calls, file_search_call, ...). Converting
		// them via the generic path would insert a spurious message between an
		// assistant tool_calls message and its tool reply, which DeepSeek rejects
		// ("insufficient tool messages following tool_calls message"). Skip them.
		if itemKind != "" && itemKind != "message" {
			bufferedReasoning = ""
			continue
		}

		bodyContent := fields["content"]
		if len(trimJSONWhitespace(bodyContent)) == 0 {
			if txt := extractRawString(fields["text"]); txt != "" {
				bodyContent, _ = json.Marshal(txt)
			}
		}
		chatContent, contentErr := convertResponsesContent(bodyContent, msgRole)
		if contentErr != nil {
			return nil, contentErr
		}
		chatMsgs = append(chatMsgs, ChatMessage{Role: msgRole, Content: chatContent})
		// Reasoning only survives across an assistant text message.
		if msgRole != "assistant" {
			bufferedReasoning = ""
		}
	}

	return chatMsgs, nil
}

// reorderToolCallMessages is the single place that enforces the tool-call
// invariant the DeepSeek / OpenAI Chat Completions schema requires: an assistant
// message with tool_calls must be immediately followed by one tool message per
// tool_call_id, in order, with nothing in between.
//
// Codex histories violate this in several ways that the builder alone can't fix:
//   - a non-tool message lands between an assistant tool_calls message and its
//     tool replies (e.g. an "Approved command prefix saved" system notice codex
//     injects mid tool-execution);
//   - a parallel tool_call's sibling output never arrives, or a call is left
//     dangling by a mid-execution reconnect (unanswered tool_call);
//   - a tool reply has no announcing assistant tool_call (orphan).
//
// It rebuilds the sequence so each assistant's answered tool_calls are followed
// directly by their replies (in call order); unanswered tool_calls are dropped
// (and an assistant left with neither tool_calls nor content is dropped); orphan
// tool replies and intervening messages are emitted in their natural position
// but never between an assistant tool_calls message and its replies.
func reorderToolCallMessages(chatMsgs []ChatMessage) []ChatMessage {
	// Index every tool reply by its tool_call_id (last occurrence wins on duplicates).
	replyIndex := make(map[string]ChatMessage)
	for _, msg := range chatMsgs {
		if msg.Role == "tool" && msg.ToolCallID != "" {
			replyIndex[msg.ToolCallID] = msg
		}
	}

	reordered := make([]ChatMessage, 0, len(chatMsgs))
	for _, msg := range chatMsgs {
		switch {
		case msg.Role == "tool":
			// A bare tool message with no tool_call_id is a direct Chat
			// Completions passthrough; keep it in place. A tool reply whose id is
			// announced by an assistant is emitted right after that assistant
			// (skip the standalone occurrence). Any other tool reply is an orphan
			// and is dropped.
			if msg.ToolCallID == "" {
				reordered = append(reordered, msg)
			}
			continue
		case len(msg.ToolCalls) > 0:
			answeredCalls := make([]ChatToolCall, 0, len(msg.ToolCalls))
			for _, tc := range msg.ToolCalls {
				if tc.ID == "" {
					continue
				}
				if _, hasReply := replyIndex[tc.ID]; hasReply {
					answeredCalls = append(answeredCalls, tc)
				}
			}
			if len(answeredCalls) == 0 {
				// No answered tool_calls left: keep as a plain message if it has
				// content, otherwise drop it entirely.
				if hasNoUsableContent(msg.Content) {
					continue
				}
				msg.ToolCalls = nil
				reordered = append(reordered, msg)
				continue
			}
			msg.ToolCalls = answeredCalls
			reordered = append(reordered, msg)
			for _, tc := range answeredCalls {
				reordered = append(reordered, replyIndex[tc.ID])
			}
		default:
			reordered = append(reordered, msg)
		}
	}
	return reordered
}

// hasNoUsableContent reports whether a chat message content holds no usable text.
func hasNoUsableContent(raw json.RawMessage) bool {
	raw = trimJSONWhitespace(raw)
	if len(raw) == 0 || string(raw) == "null" || string(raw) == `""` {
		return true
	}
	return extractChatMessageText(raw) == ""
}

// gatherReasoningText pulls the reasoning text out of a Responses
// reasoning item. The Chat->Responses bridge writes the upstream reasoning_content
// verbatim into the summary_text parts (see finalizeReasoningItem), so codex
// round-trips it there; prefer summary[].text and fall back to content.
func gatherReasoningText(fields map[string]json.RawMessage) string {
	var fragments []string
	extractParts := func(raw json.RawMessage) {
		raw = trimJSONWhitespace(raw)
		if len(raw) == 0 || string(raw) == "null" {
			return
		}
		var partArray []map[string]json.RawMessage
		if unmarshalErr := json.Unmarshal(raw, &partArray); unmarshalErr == nil {
			for _, p := range partArray {
				if txt := extractRawString(p["text"]); txt != "" {
					fragments = append(fragments, txt)
				}
			}
			return
		}
		if txt := extractRawString(raw); txt != "" {
			fragments = append(fragments, txt)
		}
	}
	extractParts(fields["summary"])
	if len(fragments) == 0 {
		extractParts(fields["content"])
	}
	return strings.Join(fragments, "\n")
}

// mapResponsesRole converts a Responses API role to its Chat Completions equivalent.
func mapResponsesRole(role string) string {
	cleaned := strings.TrimSpace(role)
	if cleaned == "" {
		return "user"
	}
	if strings.EqualFold(cleaned, "developer") {
		return "system"
	}
	return role
}

// convertResponsesContent transforms Responses content into Chat Completions content.
func convertResponsesContent(raw json.RawMessage, role string) (json.RawMessage, error) {
	raw = trimJSONWhitespace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		placeholder, _ := json.Marshal("")
		return placeholder, nil
	}

	var plainStr string
	if unmarshalErr := json.Unmarshal(raw, &plainStr); unmarshalErr == nil {
		return raw, nil
	}

	var partArray []json.RawMessage
	if unmarshalErr := json.Unmarshal(raw, &partArray); unmarshalErr == nil {
		return convertResponsesContentParts(partArray, role)
	}

	var singleObj map[string]json.RawMessage
	if unmarshalErr := json.Unmarshal(raw, &singleObj); unmarshalErr == nil {
		return convertSingleResponsesPart(extractRawString(singleObj["type"]), singleObj)
	}

	return raw, nil
}

// convertResponsesContentParts converts an array of Responses content parts to Chat format.
func convertResponsesContentParts(partArray []json.RawMessage, role string) (json.RawMessage, error) {
	var textSegments []string
	var chatSegments []ChatContentPart
	containsMedia := false

	for _, rawPart := range partArray {
		var fields map[string]json.RawMessage
		if parseErr := json.Unmarshal(rawPart, &fields); parseErr != nil {
			continue
		}
		partKind := extractRawString(fields["type"])
		switch partKind {
		case "input_text", "output_text", "text", "":
			txt := extractRawString(fields["text"])
			if txt == "" {
				continue
			}
			textSegments = append(textSegments, txt)
			chatSegments = append(chatSegments, ChatContentPart{Type: "text", Text: txt})
		case "input_image", "image_url":
			imgURL := extractRawString(fields["image_url"])
			if imgURL == "" {
				imgURL = extractNestedString(fields["image_url"], "url")
			}
			if imgURL == "" {
				continue
			}
			containsMedia = true
			chatSegments = append(chatSegments, ChatContentPart{
				Type:     "image_url",
				ImageURL: &ChatImageURL{URL: imgURL},
			})
		}
	}

	if !containsMedia {
		merged, _ := json.Marshal(strings.Join(textSegments, "\n\n"))
		return merged, nil
	}
	if role != "user" {
		merged, _ := json.Marshal(strings.Join(textSegments, "\n\n"))
		return merged, nil
	}
	if len(chatSegments) == 0 {
		placeholder, _ := json.Marshal("")
		return placeholder, nil
	}
	return json.Marshal(chatSegments)
}

// convertSingleResponsesPart converts a single Responses content part to Chat format.
func convertSingleResponsesPart(partKind string, fields map[string]json.RawMessage) (json.RawMessage, error) {
	switch partKind {
	case "input_image", "image_url":
		imgURL := extractRawString(fields["image_url"])
		if imgURL == "" {
			imgURL = extractNestedString(fields["image_url"], "url")
		}
		return json.Marshal([]ChatContentPart{{
			Type:     "image_url",
			ImageURL: &ChatImageURL{URL: imgURL},
		}})
	default:
		return json.Marshal(extractRawString(fields["text"]))
	}
}

// convertResponsesTools converts Responses tools to Chat Completions tools.
func convertResponsesTools(tools []ResponsesTool) []ChatTool {
	chatTools := make([]ChatTool, 0, len(tools))
	for _, t := range tools {
		if t.Type != "function" {
			continue
		}
		chatTools = append(chatTools, ChatTool{
			Type: "function",
			Function: &ChatFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
				Strict:      t.Strict,
			},
		})
	}
	return chatTools
}

// convertResponsesToolChoice converts Responses tool choice to Chat Completions format.
func convertResponsesToolChoice(raw json.RawMessage) json.RawMessage {
	var choiceMap map[string]json.RawMessage
	if parseErr := json.Unmarshal(raw, &choiceMap); parseErr != nil {
		return raw
	}
	if extractRawString(choiceMap["type"]) != "function" {
		return raw
	}
	fnName := extractRawString(choiceMap["name"])
	if fnName == "" {
		fnName = extractNestedString(choiceMap["function"], "name")
	}
	if fnName == "" {
		return raw
	}
	converted, marshalErr := json.Marshal(map[string]any{
		"type": "function",
		"function": map[string]string{
			"name": fnName,
		},
	})
	if marshalErr != nil {
		return raw
	}
	return converted
}

// ChatCompletionsResponseToResponses converts a non-streaming Chat Completions
// response into a Responses API response.
func ChatCompletionsResponseToResponses(resp *ChatCompletionsResponse, model string) *ResponsesResponse {
	responseID := ""
	if resp != nil {
		responseID = resp.ID
	}
	if responseID == "" {
		responseID = generateResponsesID()
	}

	apiResp := &ResponsesResponse{
		ID:     responseID,
		Object: "response",
		Model:  model,
		Status: "completed",
	}
	if resp == nil {
		apiResp.Output = []ResponsesOutput{buildEmptyMessageOutput()}
		return apiResp
	}
	if apiResp.Model == "" {
		apiResp.Model = resp.Model
	}

	if len(resp.Choices) > 0 {
		primary := resp.Choices[0]
		apiResp.Output = convertChatMessageToOutput(primary.Message)
		if primary.FinishReason == "length" {
			apiResp.Status = "incomplete"
			apiResp.IncompleteDetails = &ResponsesIncompleteDetails{Reason: "max_output_tokens"}
		}
	}
	if len(apiResp.Output) == 0 {
		apiResp.Output = []ResponsesOutput{buildEmptyMessageOutput()}
	}
	if resp.Usage != nil {
		apiResp.Usage = ChatUsageToResponsesUsage(resp.Usage)
	}
	return apiResp
}

// convertChatMessageToOutput converts a Chat message into Responses output items.
func convertChatMessageToOutput(msg ChatMessage) []ResponsesOutput {
	var outputItems []ResponsesOutput
	if msg.ReasoningContent != "" {
		outputItems = append(outputItems, ResponsesOutput{
			Type: "reasoning",
			ID:   generateItemID(),
			Summary: []ResponsesSummary{{
				Type: "summary_text",
				Text: msg.ReasoningContent,
			}},
		})
	}

	bodyText := extractChatMessageText(msg.Content)
	if bodyText == "" && strings.TrimSpace(msg.ReasoningContent) != "" && len(msg.ToolCalls) == 0 {
		bodyText = msg.ReasoningContent
	}
	if bodyText != "" || len(msg.ToolCalls) == 0 {
		outputItems = append(outputItems, ResponsesOutput{
			Type: "message",
			ID:   generateItemID(),
			Role: "assistant",
			Content: []ResponsesContentPart{{
				Type: "output_text",
				Text: bodyText,
			}},
			Status: "completed",
		})
	}

	for _, tc := range msg.ToolCalls {
		fnArgs := tc.Function.Arguments
		if strings.TrimSpace(fnArgs) == "" {
			fnArgs = "{}"
		}
		outputItems = append(outputItems, ResponsesOutput{
			Type:      "function_call",
			ID:        generateItemID(),
			CallID:    tc.ID,
			Name:      tc.Function.Name,
			Arguments: fnArgs,
			Status:    "completed",
		})
	}

	return outputItems
}

// buildEmptyMessageOutput creates an empty assistant message output.
func buildEmptyMessageOutput() ResponsesOutput {
	return ResponsesOutput{
		Type:    "message",
		ID:      generateItemID(),
		Role:    "assistant",
		Content: []ResponsesContentPart{{Type: "output_text", Text: ""}},
		Status:  "completed",
	}
}

// extractChatMessageText extracts the text content from a Chat message's content field.
func extractChatMessageText(raw json.RawMessage) string {
	raw = trimJSONWhitespace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var plainStr string
	if unmarshalErr := json.Unmarshal(raw, &plainStr); unmarshalErr == nil {
		return plainStr
	}
	var contentParts []ChatContentPart
	if unmarshalErr := json.Unmarshal(raw, &contentParts); unmarshalErr == nil {
		var textPieces []string
		for _, cp := range contentParts {
			if cp.Type == "text" && cp.Text != "" {
				textPieces = append(textPieces, cp.Text)
			}
		}
		return strings.Join(textPieces, "\n\n")
	}
	return ""
}

// ChatUsageToResponsesUsage converts Chat Completions token usage to Responses
// usage shape.
func ChatUsageToResponsesUsage(usage *ChatUsage) *ResponsesUsage {
	if usage == nil {
		return nil
	}
	converted := &ResponsesUsage{
		InputTokens:  usage.PromptTokens,
		OutputTokens: usage.CompletionTokens,
		TotalTokens:  usage.TotalTokens,
	}
	if converted.TotalTokens == 0 {
		converted.TotalTokens = converted.InputTokens + converted.OutputTokens
	}
	if usage.PromptTokensDetails != nil && usage.PromptTokensDetails.CachedTokens > 0 {
		converted.InputTokensDetails = &ResponsesInputTokensDetails{
			CachedTokens: usage.PromptTokensDetails.CachedTokens,
		}
	}
	return converted
}

// ChatCompletionsToResponsesStreamState tracks state while converting Chat
// Completions SSE chunks into Responses SSE events.
type ChatCompletionsToResponsesStreamState struct {
	ResponseID     string
	Model          string
	Created        int64
	SequenceNumber int
	CreatedSent    bool
	CompletedSent  bool

	// nextOutputIndex assigns sequential output_index values to items as they
	// are opened (reasoning, message, tool calls), so the streamed indices match
	// the order of items in the final response.output array.
	nextOutputIndex int

	// Reasoning item lifecycle. DeepSeek-style upstreams stream all
	// reasoning_content before any content, so reasoning is modeled as its own
	// "reasoning" output item that must be opened (output_item.added) before any
	// reasoning delta and closed before the message/tool items open.
	ReasoningItemID string
	ReasoningIndex  int
	ReasoningOpen   bool
	ReasoningDone   bool

	// Message item + output_text content-part lifecycle.
	MessageItemID string
	MessageIndex  int
	TextPartOpen  bool

	Text      strings.Builder
	Reasoning strings.Builder

	// Tool-call lifecycle, keyed by the upstream tool_call index.
	ToolCalls       map[int]*ChatToolCall
	ToolItemIDs     map[int]string
	ToolOutputIndex map[int]int

	FinishReason string
	Usage        *ResponsesUsage
}

// NewChatCompletionsToResponsesStreamState returns an initialized stream state.
func NewChatCompletionsToResponsesStreamState(model string) *ChatCompletionsToResponsesStreamState {
	return &ChatCompletionsToResponsesStreamState{
		ResponseID:      generateResponsesID(),
		Model:           model,
		Created:         time.Now().Unix(),
		ToolCalls:       make(map[int]*ChatToolCall),
		ToolItemIDs:     make(map[int]string),
		ToolOutputIndex: make(map[int]int),
	}
}

func (st *ChatCompletionsToResponsesStreamState) allocOutputIndex() int {
	idx := st.nextOutputIndex
	st.nextOutputIndex++
	return idx
}

// ChatCompletionsChunkToResponsesEvents converts one Chat Completions stream
// chunk into zero or more Responses stream events.
func ChatCompletionsChunkToResponsesEvents(
	chunk *ChatCompletionsChunk,
	st *ChatCompletionsToResponsesStreamState,
) []ResponsesStreamEvent {
	if chunk == nil || st == nil {
		return nil
	}
	if chunk.ID != "" {
		st.ResponseID = chunk.ID
	}
	if st.Model == "" && chunk.Model != "" {
		st.Model = chunk.Model
	}
	if chunk.Usage != nil {
		st.Usage = ChatUsageToResponsesUsage(chunk.Usage)
	}

	var emitted []ResponsesStreamEvent
	emitted = append(emitted, emitCreatedIfNeeded(st)...)

	for _, ch := range chunk.Choices {
		// Reasoning is emitted as its own output item and must be opened
		// (output_item.added + reasoning_summary_part.added) before the first
		// delta, otherwise a strict client discards the delta. The leading
		// empty-string reasoning delta upstreams send is filtered out.
		if ch.Delta.ReasoningContent != nil && *ch.Delta.ReasoningContent != "" {
			emitted = append(emitted, openReasoningItemIfNeeded(st)...)
			_, _ = st.Reasoning.WriteString(*ch.Delta.ReasoningContent)
			emitted = append(emitted, buildStreamEvent(st, "response.reasoning_summary_text.delta", &ResponsesStreamEvent{
				OutputIndex:  st.ReasoningIndex,
				SummaryIndex: 0,
				Delta:        *ch.Delta.ReasoningContent,
				ItemID:       st.ReasoningItemID,
			}))
		}
		if ch.Delta.Content != nil && *ch.Delta.Content != "" {
			// First real content closes the reasoning item, then opens the
			// message item and its output_text content part.
			emitted = append(emitted, finalizeReasoningItem(st)...)
			emitted = append(emitted, openMessageItemIfNeeded(st)...)
			emitted = append(emitted, openTextPartIfNeeded(st)...)
			_, _ = st.Text.WriteString(*ch.Delta.Content)
			emitted = append(emitted, buildStreamEvent(st, "response.output_text.delta", &ResponsesStreamEvent{
				OutputIndex:  st.MessageIndex,
				ContentIndex: 0,
				Delta:        *ch.Delta.Content,
				ItemID:       st.MessageItemID,
			}))
		}
		for _, toolDelta := range ch.Delta.ToolCalls {
			tcIdx := 0
			if toolDelta.Index != nil {
				tcIdx = *toolDelta.Index
			}
			existing, tracked := st.ToolCalls[tcIdx]
			if !tracked {
				// A tool call closes any open reasoning item first.
				emitted = append(emitted, finalizeReasoningItem(st)...)
				copied := toolDelta
				if copied.ID == "" {
					copied.ID = generateItemID()
				}
				copied.Type = "function"
				st.ToolCalls[tcIdx] = &copied
				existing = &copied
				newItemID := generateItemID()
				st.ToolItemIDs[tcIdx] = newItemID
				st.ToolOutputIndex[tcIdx] = st.allocOutputIndex()
				emitted = append(emitted, buildStreamEvent(st, "response.output_item.added", &ResponsesStreamEvent{
					OutputIndex: st.ToolOutputIndex[tcIdx],
					Item: &ResponsesOutput{
						Type:   "function_call",
						ID:     newItemID,
						CallID: existing.ID,
						Name:   existing.Function.Name,
						Status: "in_progress",
					},
				}))
			} else {
				if toolDelta.ID != "" {
					existing.ID = toolDelta.ID
				}
				if toolDelta.Function.Name != "" {
					existing.Function.Name = toolDelta.Function.Name
				}
			}
			if toolDelta.Function.Arguments != "" {
				existing.Function.Arguments += toolDelta.Function.Arguments
				emitted = append(emitted, buildStreamEvent(st, "response.function_call_arguments.delta", &ResponsesStreamEvent{
					OutputIndex: st.ToolOutputIndex[tcIdx],
					ItemID:      st.ToolItemIDs[tcIdx],
					Delta:       toolDelta.Function.Arguments,
					CallID:      existing.ID,
					Name:        existing.Function.Name,
				}))
			}
		}
		if ch.FinishReason != nil && *ch.FinishReason != "" {
			st.FinishReason = *ch.FinishReason
		}
	}

	return emitted
}

// FinalizeChatCompletionsResponsesStream emits terminal Responses events.
func FinalizeChatCompletionsResponsesStream(st *ChatCompletionsToResponsesStreamState) []ResponsesStreamEvent {
	if st == nil || st.CompletedSent {
		return nil
	}
	var emitted []ResponsesStreamEvent
	emitted = append(emitted, emitCreatedIfNeeded(st)...)

	// Close a reasoning item that never transitioned to content (reasoning-only
	// or empty completion).
	emitted = append(emitted, finalizeReasoningItem(st)...)
	emitted = append(emitted, emitReasoningFallbackMessage(st)...)

	if st.MessageItemID != "" {
		if st.TextPartOpen {
			emitted = append(emitted, buildStreamEvent(st, "response.output_text.done", &ResponsesStreamEvent{
				OutputIndex:  st.MessageIndex,
				ContentIndex: 0,
				Text:         st.Text.String(),
				ItemID:       st.MessageItemID,
			}))
			emitted = append(emitted, buildStreamEvent(st, "response.content_part.done", &ResponsesStreamEvent{
				OutputIndex:  st.MessageIndex,
				ContentIndex: 0,
				ItemID:       st.MessageItemID,
				Part:         &ResponsesContentPart{Type: "output_text", Text: st.Text.String()},
			}))
		}
		emitted = append(emitted, buildStreamEvent(st, "response.output_item.done", &ResponsesStreamEvent{
			OutputIndex: st.MessageIndex,
			Item: &ResponsesOutput{
				Type:    "message",
				ID:      st.MessageItemID,
				Role:    "assistant",
				Content: []ResponsesContentPart{{Type: "output_text", Text: st.Text.String()}},
				Status:  "completed",
			},
		}))
	}

	// Close every function_call item opened during the stream. Codex finalizes a
	// tool call only after function_call_arguments.done + output_item.done for
	// that item; without them the call never completes and the session wedges.
	emitted = append(emitted, finalizeAllToolItems(st)...)

	completionStatus := "completed"
	var incompleteInfo *ResponsesIncompleteDetails
	if st.FinishReason == "length" {
		completionStatus = "incomplete"
		incompleteInfo = &ResponsesIncompleteDetails{Reason: "max_output_tokens"}
	}

	st.CompletedSent = true
	emitted = append(emitted, buildStreamEvent(st, "response.completed", &ResponsesStreamEvent{
		Response: &ResponsesResponse{
			ID:                st.ResponseID,
			Object:            "response",
			Model:             st.Model,
			Status:            completionStatus,
			Output:            st.buildFinalOutput(),
			Usage:             st.Usage,
			IncompleteDetails: incompleteInfo,
		},
	}))
	return emitted
}

// emitCreatedIfNeeded emits the response.created event if not yet sent.
func emitCreatedIfNeeded(st *ChatCompletionsToResponsesStreamState) []ResponsesStreamEvent {
	if st.CreatedSent {
		return nil
	}
	st.CreatedSent = true
	return []ResponsesStreamEvent{buildStreamEvent(st, "response.created", &ResponsesStreamEvent{
		Response: &ResponsesResponse{
			ID:     st.ResponseID,
			Object: "response",
			Model:  st.Model,
			Status: "in_progress",
			Output: []ResponsesOutput{},
		},
	})}
}

// openReasoningItemIfNeeded opens the reasoning output item (output_item.added +
// reasoning_summary_part.added) before the first reasoning delta.
func openReasoningItemIfNeeded(st *ChatCompletionsToResponsesStreamState) []ResponsesStreamEvent {
	if st.ReasoningOpen || st.ReasoningDone {
		return nil
	}
	st.ReasoningOpen = true
	st.ReasoningItemID = generateItemID()
	st.ReasoningIndex = st.allocOutputIndex()
	return []ResponsesStreamEvent{
		buildStreamEvent(st, "response.output_item.added", &ResponsesStreamEvent{
			OutputIndex: st.ReasoningIndex,
			Item:        &ResponsesOutput{Type: "reasoning", ID: st.ReasoningItemID, Status: "in_progress"},
		}),
		buildStreamEvent(st, "response.reasoning_summary_part.added", &ResponsesStreamEvent{
			OutputIndex:  st.ReasoningIndex,
			SummaryIndex: 0,
			ItemID:       st.ReasoningItemID,
			Part:         &ResponsesContentPart{Type: "summary_text"},
		}),
	}
}

// finalizeReasoningItem emits the reasoning item's terminal events
// (reasoning_summary_text.done + reasoning_summary_part.done + output_item.done).
func finalizeReasoningItem(st *ChatCompletionsToResponsesStreamState) []ResponsesStreamEvent {
	if !st.ReasoningOpen {
		return nil
	}
	st.ReasoningOpen = false
	st.ReasoningDone = true
	reasoningText := st.Reasoning.String()
	return []ResponsesStreamEvent{
		buildStreamEvent(st, "response.reasoning_summary_text.done", &ResponsesStreamEvent{
			OutputIndex:  st.ReasoningIndex,
			SummaryIndex: 0,
			Text:         reasoningText,
			ItemID:       st.ReasoningItemID,
		}),
		buildStreamEvent(st, "response.reasoning_summary_part.done", &ResponsesStreamEvent{
			OutputIndex:  st.ReasoningIndex,
			SummaryIndex: 0,
			ItemID:       st.ReasoningItemID,
			Part:         &ResponsesContentPart{Type: "summary_text", Text: reasoningText},
		}),
		buildStreamEvent(st, "response.output_item.done", &ResponsesStreamEvent{
			OutputIndex: st.ReasoningIndex,
			Item: &ResponsesOutput{
				Type:    "reasoning",
				ID:      st.ReasoningItemID,
				Status:  "completed",
				Summary: []ResponsesSummary{{Type: "summary_text", Text: reasoningText}},
			},
		}),
	}
}

// emitReasoningFallbackMessage synthesizes a message item from reasoning content
// when no regular content or tool calls were produced.
func emitReasoningFallbackMessage(st *ChatCompletionsToResponsesStreamState) []ResponsesStreamEvent {
	if st == nil ||
		st.MessageItemID != "" ||
		st.Text.Len() > 0 ||
		st.Reasoning.Len() == 0 ||
		len(st.ToolCalls) > 0 {
		return nil
	}

	reasoningBody := st.Reasoning.String()
	if strings.TrimSpace(reasoningBody) == "" {
		return nil
	}

	var emitted []ResponsesStreamEvent
	emitted = append(emitted, openMessageItemIfNeeded(st)...)
	emitted = append(emitted, openTextPartIfNeeded(st)...)
	_, _ = st.Text.WriteString(reasoningBody)
	emitted = append(emitted, buildStreamEvent(st, "response.output_text.delta", &ResponsesStreamEvent{
		OutputIndex:  st.MessageIndex,
		ContentIndex: 0,
		Delta:        reasoningBody,
		ItemID:       st.MessageItemID,
	}))
	return emitted
}

// openMessageItemIfNeeded opens the message output item if not yet opened.
func openMessageItemIfNeeded(st *ChatCompletionsToResponsesStreamState) []ResponsesStreamEvent {
	if st.MessageItemID != "" {
		return nil
	}
	st.MessageItemID = generateItemID()
	st.MessageIndex = st.allocOutputIndex()
	return []ResponsesStreamEvent{buildStreamEvent(st, "response.output_item.added", &ResponsesStreamEvent{
		OutputIndex: st.MessageIndex,
		Item: &ResponsesOutput{
			Type:    "message",
			ID:      st.MessageItemID,
			Role:    "assistant",
			Status:  "in_progress",
			Content: []ResponsesContentPart{{Type: "output_text"}},
		},
	})}
}

// openTextPartIfNeeded opens the output_text content part if not yet opened.
func openTextPartIfNeeded(st *ChatCompletionsToResponsesStreamState) []ResponsesStreamEvent {
	if st.TextPartOpen {
		return nil
	}
	st.TextPartOpen = true
	return []ResponsesStreamEvent{buildStreamEvent(st, "response.content_part.added", &ResponsesStreamEvent{
		OutputIndex:  st.MessageIndex,
		ContentIndex: 0,
		ItemID:       st.MessageItemID,
		Part:         &ResponsesContentPart{Type: "output_text", Text: ""},
	})}
}

// finalizeAllToolItems emits function_call_arguments.done + output_item.done for
// every tool call opened during the stream.
func finalizeAllToolItems(st *ChatCompletionsToResponsesStreamState) []ResponsesStreamEvent {
	if len(st.ToolCalls) == 0 {
		return nil
	}
	var emitted []ResponsesStreamEvent
	for tcIdx := 0; tcIdx < len(st.ToolCalls); tcIdx++ {
		tc, exists := st.ToolCalls[tcIdx]
		if !exists || tc == nil {
			continue
		}
		itemID, wasOpened := st.ToolItemIDs[tcIdx]
		if !wasOpened {
			continue
		}
		fnArgs := tc.Function.Arguments
		if strings.TrimSpace(fnArgs) == "" {
			fnArgs = "{}"
		}
		outIdx := st.ToolOutputIndex[tcIdx]
		emitted = append(emitted,
			buildStreamEvent(st, "response.function_call_arguments.done", &ResponsesStreamEvent{
				OutputIndex: outIdx,
				ItemID:      itemID,
				CallID:      tc.ID,
				Name:        tc.Function.Name,
				Arguments:   fnArgs,
			}),
			buildStreamEvent(st, "response.output_item.done", &ResponsesStreamEvent{
				OutputIndex: outIdx,
				Item: &ResponsesOutput{
					Type:      "function_call",
					ID:        itemID,
					CallID:    tc.ID,
					Name:      tc.Function.Name,
					Arguments: fnArgs,
					Status:    "completed",
				},
			}),
		)
	}
	return emitted
}

// buildFinalOutput assembles the complete output array for the final response.completed event.
func (st *ChatCompletionsToResponsesStreamState) buildFinalOutput() []ResponsesOutput {
	var outputItems []ResponsesOutput
	if st.Reasoning.Len() > 0 {
		outputItems = append(outputItems, ResponsesOutput{
			Type: "reasoning",
			ID:   generateItemID(),
			Summary: []ResponsesSummary{{
				Type: "summary_text",
				Text: st.Reasoning.String(),
			}},
		})
	}
	if st.MessageItemID != "" || len(st.ToolCalls) == 0 {
		outputItems = append(outputItems, ResponsesOutput{
			Type: "message",
			ID:   nonEmpty(st.MessageItemID, generateItemID()),
			Role: "assistant",
			Content: []ResponsesContentPart{{
				Type: "output_text",
				Text: st.Text.String(),
			}},
			Status: "completed",
		})
	}
	for tcIdx := 0; tcIdx < len(st.ToolCalls); tcIdx++ {
		tc, exists := st.ToolCalls[tcIdx]
		if !exists || tc == nil {
			continue
		}
		fnArgs := tc.Function.Arguments
		if strings.TrimSpace(fnArgs) == "" {
			fnArgs = "{}"
		}
		outputItems = append(outputItems, ResponsesOutput{
			Type:      "function_call",
			ID:        generateItemID(),
			CallID:    tc.ID,
			Name:      tc.Function.Name,
			Arguments: fnArgs,
			Status:    "completed",
		})
	}
	return outputItems
}

// buildStreamEvent creates a numbered ResponsesStreamEvent.
func buildStreamEvent(
	st *ChatCompletionsToResponsesStreamState,
	eventKind string,
	template *ResponsesStreamEvent,
) ResponsesStreamEvent {
	seqNum := st.SequenceNumber
	st.SequenceNumber++
	evt := *template
	evt.Type = eventKind
	evt.SequenceNumber = seqNum
	return evt
}

// extractRawString extracts a plain string from a JSON-encoded value.
func extractRawString(raw json.RawMessage) string {
	raw = trimJSONWhitespace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var decoded string
	if unmarshalErr := json.Unmarshal(raw, &decoded); unmarshalErr == nil {
		return decoded
	}
	return ""
}

// extractNestedString extracts a string from a nested JSON object field.
func extractNestedString(raw json.RawMessage, fieldName string) string {
	var nested map[string]json.RawMessage
	if parseErr := json.Unmarshal(raw, &nested); parseErr != nil {
		return ""
	}
	return extractRawString(nested[fieldName])
}

// trimJSONWhitespace strips leading and trailing whitespace from JSON bytes.
func trimJSONWhitespace(raw json.RawMessage) json.RawMessage {
	return json.RawMessage(strings.TrimSpace(string(raw)))
}

func nonEmpty(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

// responsesInputToChatMessages is a compatibility alias retained for test callers.
func responsesInputToChatMessages(instructions string, rawInput json.RawMessage) ([]ChatMessage, error) {
	return transformInputToChatMessages(instructions, rawInput)
}
