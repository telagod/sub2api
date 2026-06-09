package service

import (
	"encoding/json"

	"github.com/telagod/subme/internal/pkg/apicompat"
)

const openAICompatAnthropicReplayMaxTailMessages = 12

func applyAnthropicCompatFullReplayGuard(req *apicompat.AnthropicRequest) bool {
	if req == nil || len(req.Messages) <= openAICompatAnthropicReplayMaxTailMessages {
		return false
	}

	cutoff := len(req.Messages) - openAICompatAnthropicReplayMaxTailMessages
	cutoff = expandAnthropicCompatTrimBoundary(req.Messages, cutoff)
	if cutoff <= 0 {
		return false
	}

	trimmed := make([]apicompat.AnthropicMessage, len(req.Messages)-cutoff)
	copy(trimmed, req.Messages[cutoff:])
	req.Messages = trimmed
	return true
}

func expandAnthropicCompatTrimBoundary(msgs []apicompat.AnthropicMessage, boundary int) int {
	if boundary <= 0 || boundary >= len(msgs) {
		return boundary
	}

	usePositions := make(map[string]int)
	resultPositions := make(map[string]int)
	for idx, m := range msgs {
		toolUses, toolResults := anthropicCompatMessageToolIDs(m)
		for _, tid := range toolUses {
			if _, found := usePositions[tid]; !found {
				usePositions[tid] = idx
			}
		}
		for _, tid := range toolResults {
			if _, found := resultPositions[tid]; !found {
				resultPositions[tid] = idx
			}
		}
	}

	for {
		expanded := boundary
		for pos := boundary; pos < len(msgs); pos++ {
			tUses, tResults := anthropicCompatMessageToolIDs(msgs[pos])
			for _, tid := range tResults {
				if uPos, ok := usePositions[tid]; ok && uPos < expanded {
					expanded = uPos
				}
			}
			for _, tid := range tUses {
				if rPos, ok := resultPositions[tid]; ok && rPos < expanded {
					expanded = rPos
				}
			}
		}
		if expanded == boundary {
			break
		}
		boundary = expanded
	}
	return boundary
}

func anthropicCompatMessageToolIDs(m apicompat.AnthropicMessage) ([]string, []string) {
	var contentBlocks []apicompat.AnthropicContentBlock
	if unmarshalErr := json.Unmarshal(m.Content, &contentBlocks); unmarshalErr != nil {
		return nil, nil
	}

	toolUseIDs := make([]string, 0, 1)
	toolResultIDs := make([]string, 0, 1)
	for _, blk := range contentBlocks {
		if blk.Type == "tool_use" && blk.ID != "" {
			toolUseIDs = append(toolUseIDs, blk.ID)
		} else if blk.Type == "tool_result" && blk.ToolUseID != "" {
			toolResultIDs = append(toolResultIDs, blk.ToolUseID)
		}
	}
	return toolUseIDs, toolResultIDs
}
