package apicompat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// prompt_cache_key is a cache-routing hint: clients that send it on
// /v1/chat/completions expect it to survive the CC→Responses conversion,
// otherwise upstream prompt-cache hit rates silently degrade.
// Upstream-Ref: d251487d (behavior-aligned, independent implementation)
func TestChatCompletionsToResponses_PropagatesPromptCacheKey(t *testing.T) {
	body := []byte(`{
		"model": "gpt-5.2",
		"prompt_cache_key": "tenant-42-session-abc",
		"messages": [{"role": "user", "content": "hello"}]
	}`)

	var req ChatCompletionsRequest
	require.NoError(t, json.Unmarshal(body, &req))
	require.Equal(t, "tenant-42-session-abc", req.PromptCacheKey)

	out, err := ChatCompletionsToResponses(&req)
	require.NoError(t, err)
	require.Equal(t, "tenant-42-session-abc", out.PromptCacheKey)

	// The marshaled upstream body must carry the field end-to-end.
	marshaled, err := json.Marshal(out)
	require.NoError(t, err)
	var roundTrip map[string]any
	require.NoError(t, json.Unmarshal(marshaled, &roundTrip))
	require.Equal(t, "tenant-42-session-abc", roundTrip["prompt_cache_key"])
}

func TestChatCompletionsToResponses_OmitsEmptyPromptCacheKey(t *testing.T) {
	req := &ChatCompletionsRequest{
		Model:    "gpt-5.2",
		Messages: []ChatMessage{{Role: "user", Content: json.RawMessage(`"hi"`)}},
	}

	out, err := ChatCompletionsToResponses(req)
	require.NoError(t, err)
	require.Empty(t, out.PromptCacheKey)

	marshaled, err := json.Marshal(out)
	require.NoError(t, err)
	var roundTrip map[string]any
	require.NoError(t, json.Unmarshal(marshaled, &roundTrip))
	require.NotContains(t, roundTrip, "prompt_cache_key")
}
