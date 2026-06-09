package service

import (
	"bytes"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

const openAICompatMessagesBridgeContextKey = "openai_compat_messages_bridge"

func isOpenAICompatMessagesBridgeBody(payload []byte) bool {
	if len(payload) == 0 {
		return false
	}
	if bytes.Contains(payload, []byte(openAICompatClaudeCodeTodoGuardMarker)) {
		return true
	}
	cacheKey := gjson.GetBytes(payload, "prompt_cache_key").String()
	return isOpenAICompatMessagesBridgePromptCacheKey(cacheKey)
}

func isOpenAICompatMessagesBridgeRequestBody(reqPayload map[string]any) bool {
	if reqPayload == nil {
		return false
	}
	if items, ok := reqPayload["input"].([]any); ok && inputContainsText(items, openAICompatClaudeCodeTodoGuardMarker) {
		return true
	}
	rawKey := firstNonEmptyString(reqPayload["prompt_cache_key"])
	return isOpenAICompatMessagesBridgePromptCacheKey(rawKey)
}

func isOpenAICompatMessagesBridgePromptCacheKey(val string) bool {
	trimmed := strings.TrimSpace(val)
	switch {
	case strings.HasPrefix(trimmed, "anthropic-metadata-"):
		return true
	case strings.HasPrefix(trimmed, "anthropic-cache-"):
		return true
	case strings.HasPrefix(trimmed, "anthropic-digest-"):
		return true
	default:
		return false
	}
}

func setOpenAICompatMessagesBridgeContext(ctx *gin.Context, flag bool) {
	if ctx == nil || !flag {
		return
	}
	ctx.Set(openAICompatMessagesBridgeContextKey, true)
}

func isOpenAICompatMessagesBridgeContext(ctx *gin.Context) bool {
	if ctx == nil {
		return false
	}
	raw, exists := ctx.Get(openAICompatMessagesBridgeContextKey)
	if !exists {
		return false
	}
	b, ok := raw.(bool)
	return ok && b
}
