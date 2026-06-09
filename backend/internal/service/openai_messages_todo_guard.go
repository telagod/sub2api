package service

import (
	"encoding/json"
	"strings"

	"github.com/telagod/subme/internal/pkg/apicompat"
)

const (
	openAICompatClaudeCodeTodoGuardMarker = "<sub2api-claude-code-todo-guard>"
	openAICompatClaudeCodeTodoGuardText   = openAICompatClaudeCodeTodoGuardMarker + "\nWhen using Claude Code todo or task tracking tools, keep the visible task list consistent. Do not send final or summary text while any item remains in_progress. Before finishing, asking the user to choose, or reporting a blocker, update the todo list so completed work is completed and deferred work is pending/open; leave an item in_progress only when active work will continue in the same turn.\n</sub2api-claude-code-todo-guard>"
)

func appendOpenAICompatClaudeCodeTodoGuard(req *apicompat.ResponsesRequest) bool {
	if req == nil || len(req.Input) == 0 {
		return false
	}

	var parsed []apicompat.ResponsesInputItem
	if unmarshalErr := json.Unmarshal(req.Input, &parsed); unmarshalErr != nil {
		return false
	}
	if len(parsed) == 0 {
		return false
	}
	if responsesInputItemsContainText(parsed, openAICompatClaudeCodeTodoGuardMarker) {
		return false
	}

	guardContent, marshalErr := json.Marshal([]apicompat.ResponsesContentPart{{
		Type: "input_text",
		Text: openAICompatClaudeCodeTodoGuardText,
	}})
	if marshalErr != nil {
		return false
	}

	guardItem := apicompat.ResponsesInputItem{
		Type:    "message",
		Role:    "developer",
		Content: guardContent,
	}

	// Find the position right after leading developer messages.
	pos := 0
	for pos < len(parsed) && parsed[pos].Type == "message" && parsed[pos].Role == "developer" {
		pos++
	}

	// Insert at pos: grow slice then shift tail.
	parsed = append(parsed, apicompat.ResponsesInputItem{})
	copy(parsed[pos+1:], parsed[pos:])
	parsed[pos] = guardItem

	encoded, encodeErr := json.Marshal(parsed)
	if encodeErr != nil {
		return false
	}
	req.Input = encoded
	return true
}

func appendOpenAICompatClaudeCodeTodoGuardToRequestBody(body map[string]any) bool {
	if body == nil {
		return false
	}

	inputSlice, ok := body["input"].([]any)
	if !ok || len(inputSlice) == 0 {
		return false
	}
	if inputContainsText(inputSlice, openAICompatClaudeCodeTodoGuardMarker) {
		return false
	}

	guardMsg := map[string]any{
		"type": "message",
		"role": "developer",
		"content": []any{
			map[string]any{
				"type": "input_text",
				"text": openAICompatClaudeCodeTodoGuardText,
			},
		},
	}

	// Advance past leading developer messages.
	pos := 0
	for pos < len(inputSlice) {
		entry, entryOK := inputSlice[pos].(map[string]any)
		if !entryOK {
			break
		}
		if strings.TrimSpace(firstNonEmptyString(entry["type"])) != "message" {
			break
		}
		if strings.TrimSpace(firstNonEmptyString(entry["role"])) != "developer" {
			break
		}
		pos++
	}

	inputSlice = append(inputSlice, nil)
	copy(inputSlice[pos+1:], inputSlice[pos:])
	inputSlice[pos] = guardMsg
	body["input"] = inputSlice
	return true
}

func responsesInputItemsContainText(entries []apicompat.ResponsesInputItem, target string) bool {
	target = strings.TrimSpace(target)
	if target == "" {
		return false
	}
	for idx := range entries {
		if strings.Contains(string(entries[idx].Content), target) {
			return true
		}
	}
	return false
}

func inputContainsText(entries []any, target string) bool {
	target = strings.TrimSpace(target)
	if target == "" {
		return false
	}
	for _, entry := range entries {
		serialized, serErr := json.Marshal(entry)
		if serErr != nil {
			continue
		}
		if strings.Contains(string(serialized), target) {
			return true
		}
	}
	return false
}
