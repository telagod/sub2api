package service

import (
	"context"
	"fmt"

	"github.com/telagod/subme/internal/pkg/claude"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// stripMessageCacheControl removes $.messages[*].content[*].cache_control.
//
// Clients (especially Claude Code) tend to pin cache_control on the "current last
// user message". After a new turn appends more messages, that formerly-last message
// becomes an intermediate one with stale cache_control, breaking prefix signature
// stability. Clearing everything and letting the proxy re-inject breakpoints
// (via addMessageCacheBreakpoints) keeps multi-turn caching reliable.
func stripMessageCacheControl(data []byte) []byte {
	msgs := gjson.GetBytes(data, "messages")
	if !msgs.IsArray() {
		return data
	}
	mi := -1
	msgs.ForEach(func(_, m gjson.Result) bool {
		mi++
		blocks := m.Get("content")
		if !blocks.IsArray() {
			return true
		}
		bi := -1
		blocks.ForEach(func(_, blk gjson.Result) bool {
			bi++
			if !blk.Get("cache_control").Exists() {
				return true
			}
			jp := fmt.Sprintf("messages.%d.content.%d.cache_control", mi, bi)
			if updated, delErr := sjson.DeleteBytes(data, jp); delErr == nil {
				data = updated
			}
			return true
		})
		return true
	})
	return data
}

// addMessageCacheBreakpoints injects two stable cache breakpoints on messages:
//  1. The last message.
//  2. When there are >= 4 messages, the second-to-last user-role message.
//
// Together with a system-prompt breakpoint and a tools[-1] breakpoint, this fills
// up to 4 breakpoints (the Anthropic maximum).
//
// TTL policy:
//   - If the target block already has cache_control.ttl, it is preserved.
//   - Otherwise {"type":"ephemeral","ttl": claude.DefaultCacheControlTTL} is written.
//
// stripMessageCacheControl should be called beforehand for idempotency.
func addMessageCacheBreakpoints(data []byte) []byte {
	msgs := gjson.GetBytes(data, "messages")
	if !msgs.IsArray() {
		return data
	}
	elements := msgs.Array()
	total := len(elements)
	if total == 0 {
		return data
	}

	data = injectCacheControlOnLastContentBlock(data, total-1, &elements[total-1])

	if total >= 4 {
		usersFound := 0
		for k := total - 1; k >= 0; k-- {
			if elements[k].Get("role").String() != "user" {
				continue
			}
			usersFound++
			if usersFound == 2 {
				data = injectCacheControlOnLastContentBlock(data, k, &elements[k])
				break
			}
		}
	}

	return data
}

// rewriteMessageCacheControlIfEnabled conditionally applies legacy message cache breakpoint rewriting.
func (s *GatewayService) rewriteMessageCacheControlIfEnabled(ctx context.Context, data []byte) []byte {
	if s == nil || !s.isRewriteMessageCacheControlEnabled(ctx) {
		return data
	}
	data = stripMessageCacheControl(data)
	return addMessageCacheBreakpoints(data)
}

func (s *GatewayService) isRewriteMessageCacheControlEnabled(ctx context.Context) bool {
	if s == nil || s.settingService == nil {
		return false
	}
	return s.settingService.IsRewriteMessageCacheControlEnabled(ctx)
}

// injectCacheControlOnLastContentBlock places a cache_control breakpoint on the
// final content block of messages[idx]. If content is a plain string, it is first
// promoted to a single-element text array.
//
// msg is a pre-fetched gjson.Result snapshot to avoid a redundant GetBytes call.
func injectCacheControlOnLastContentBlock(data []byte, idx int, msg *gjson.Result) []byte {
	contentResult := msg.Get("content")

	// String content: upgrade to array with cache_control baked in.
	if contentResult.Type == gjson.String {
		textVal := contentResult.String()
		fragment := fmt.Sprintf(
			`[{"type":"text","text":%s,"cache_control":{"type":"ephemeral","ttl":%q}}]`,
			mustJSONString(textVal), claude.DefaultCacheControlTTL,
		)
		if patched, patchErr := sjson.SetRawBytes(data, fmt.Sprintf("messages.%d.content", idx), []byte(fragment)); patchErr == nil {
			data = patched
		}
		return data
	}

	if !contentResult.IsArray() {
		return data
	}
	parts := contentResult.Array()
	if len(parts) == 0 {
		return data
	}
	tailIdx := len(parts) - 1
	tailBlock := parts[tailIdx]

	// If ttl is already explicitly set, leave it alone.
	if existing := tailBlock.Get("cache_control"); existing.Exists() && existing.Get("ttl").String() != "" {
		return data
	}

	basePath := fmt.Sprintf("messages.%d.content.%d.cache_control", idx, tailIdx)
	if tailBlock.Get("cache_control").Exists() {
		// cache_control exists but without ttl: add the ttl field.
		if patched, patchErr := sjson.SetBytes(data, basePath+".ttl", claude.DefaultCacheControlTTL); patchErr == nil {
			data = patched
		}
		return data
	}
	// No cache_control at all: write the full object.
	ccJSON := fmt.Sprintf(`{"type":"ephemeral","ttl":%q}`, claude.DefaultCacheControlTTL)
	if patched, patchErr := sjson.SetRawBytes(data, basePath, []byte(ccJSON)); patchErr == nil {
		data = patched
	}
	return data
}

// mustJSONString serializes a Go string into a valid JSON string literal (with quotes),
// for use in hand-assembled JSON passed to sjson.SetRawBytes.
func mustJSONString(s string) string {
	return fmt.Sprintf("%q", s)
}
