package service

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"sort"
	"strings"

	"github.com/telagod/subme/internal/pkg/claude"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// toolNameRewriteKey is the gin.Context key for the ToolNameRewrite mapping.
// Written during request phase, read during response phase for byte-level
// fake-name to real-name restoration.
const toolNameRewriteKey = "claude_tool_name_rewrite"

// staticToolNameRewrites are fixed prefix mappings applied to tool names.
var staticToolNameRewrites = map[string]string{
	"sessions_": "cc_sess_",
	"session_":  "cc_ses_",
}

// fakeToolNamePrefixes is the pool of human-readable prefixes used when
// dynamic tool name obfuscation is active.
var fakeToolNamePrefixes = []string{
	"analyze_", "compute_", "fetch_", "generate_", "lookup_", "modify_",
	"process_", "query_", "render_", "resolve_", "sync_", "update_",
	"validate_", "convert_", "extract_", "manage_", "monitor_", "parse_",
	"review_", "search_", "transform_", "handle_", "invoke_", "notify_",
}

// dynamicToolMapThreshold is the minimum number of tools required to enable
// dynamic name obfuscation. Fewer tools (e.g. bash/edit/read) do not need it.
const dynamicToolMapThreshold = 5

// ToolNameRewrite holds the bidirectional tool name obfuscation mapping for a single request.
//   - Forward: real -> fake, applied to the outbound request body.
//   - Reverse: fake -> real, applied to each response chunk via bytes.Replace.
//
// ReverseOrdered lists (fake, real) pairs sorted by descending fake-name length,
// preventing shorter fake names that are substrings of longer ones from being
// replaced first.
type ToolNameRewrite struct {
	Forward        map[string]string
	Reverse        map[string]string
	ReverseOrdered [][2]string
}

// buildDynamicToolMap creates deterministic fake-name mappings for the given tool names.
//
// Returns nil when len(toolNames) <= dynamicToolMapThreshold (no obfuscation needed).
// Uses fnv64a over the sorted names as a seed so identical tool sets always produce
// the same mapping within a single process.
func buildDynamicToolMap(toolNames []string) map[string]string {
	if len(toolNames) <= dynamicToolMapThreshold {
		return nil
	}

	hasher := fnv.New64a()
	for idx, nm := range toolNames {
		if idx > 0 {
			_, _ = hasher.Write([]byte{0})
		}
		_, _ = hasher.Write([]byte(nm))
	}
	src := rand.New(rand.NewSource(int64(hasher.Sum64())))

	pool := make([]string, len(fakeToolNamePrefixes))
	copy(pool, fakeToolNamePrefixes)
	src.Shuffle(len(pool), func(a, b int) { pool[a], pool[b] = pool[b], pool[a] })

	result := make(map[string]string, len(toolNames))
	for i, realName := range toolNames {
		chosenPrefix := pool[i%len(pool)]
		headLen := 3
		if len(realName) < headLen {
			headLen = len(realName)
		}
		result[realName] = fmt.Sprintf("%s%s%02d", chosenPrefix, realName[:headLen], i)
	}
	return result
}

// sanitizeToolName converts a real tool name to its fake counterpart.
// Dynamic mapping takes priority; static prefix mapping is the fallback.
func sanitizeToolName(realName string, dynMap map[string]string) string {
	if dynMap != nil {
		if alias, found := dynMap[realName]; found {
			return alias
		}
	}
	for origPrefix, newPrefix := range staticToolNameRewrites {
		if strings.HasPrefix(realName, origPrefix) {
			return newPrefix + realName[len(origPrefix):]
		}
	}
	return realName
}

// shouldMimicToolName indicates whether a tool should have its name rewritten.
// Server tools (type set to something other than "" / "function" / "custom")
// are Anthropic protocol primitives and must not be renamed.
func shouldMimicToolName(toolKind string) bool {
	return toolKind == "" || toolKind == "function" || toolKind == "custom"
}

// buildToolNameRewriteFromBody scans body's tools[*].name and builds a ToolNameRewrite.
// Returns nil when no tools need obfuscation.
// This function only reads the body; actual mutation happens in applyToolNameRewriteToBody.
func buildToolNameRewriteFromBody(payload []byte) *ToolNameRewrite {
	toolsResult := gjson.GetBytes(payload, "tools")
	if !toolsResult.IsArray() {
		return nil
	}

	eligible := make([]string, 0)
	for _, t := range toolsResult.Array() {
		if !shouldMimicToolName(t.Get("type").String()) {
			continue
		}
		nm := t.Get("name").String()
		if nm != "" {
			eligible = append(eligible, nm)
		}
	}

	dynMap := buildDynamicToolMap(eligible)

	rewrite := &ToolNameRewrite{
		Forward: make(map[string]string),
		Reverse: make(map[string]string),
	}
	for _, nm := range eligible {
		alias := sanitizeToolName(nm, dynMap)
		if alias == nm {
			continue
		}
		rewrite.Forward[nm] = alias
		rewrite.Reverse[alias] = nm
	}
	if len(rewrite.Forward) == 0 {
		return nil
	}

	rewrite.ReverseOrdered = make([][2]string, 0, len(rewrite.Reverse))
	for alias, orig := range rewrite.Reverse {
		rewrite.ReverseOrdered = append(rewrite.ReverseOrdered, [2]string{alias, orig})
	}
	sort.SliceStable(rewrite.ReverseOrdered, func(a, b int) bool {
		return len(rewrite.ReverseOrdered[a][0]) > len(rewrite.ReverseOrdered[b][0])
	})

	return rewrite
}

// applyToolNameRewriteToBody applies the pre-built ToolNameRewrite to the request body:
//   - Renames $.tools[*].name (only for mimicable tools)
//   - Renames $.tool_choice.name (when $.tool_choice.type == "tool")
//   - Renames $.messages[*].content[*].name (when type == "tool_use")
//   - Injects an ephemeral cache breakpoint on $.tools[-1]
//
// Response-side bytes.Replace restores fake names back to originals.
func applyToolNameRewriteToBody(payload []byte, rewrite *ToolNameRewrite) []byte {
	if rewrite == nil || len(rewrite.Forward) == 0 {
		return applyToolsLastCacheBreakpoint(payload)
	}

	// Rewrite tool definitions.
	toolsResult := gjson.GetBytes(payload, "tools")
	if toolsResult.IsArray() {
		ti := -1
		toolsResult.ForEach(func(_, t gjson.Result) bool {
			ti++
			if !shouldMimicToolName(t.Get("type").String()) {
				return true
			}
			nm := t.Get("name").String()
			if nm == "" {
				return true
			}
			alias, exists := rewrite.Forward[nm]
			if !exists {
				return true
			}
			if patched, patchErr := sjson.SetBytes(payload, fmt.Sprintf("tools.%d.name", ti), alias); patchErr == nil {
				payload = patched
			}
			return true
		})
	}

	// Rewrite tool_choice.
	if tc := gjson.GetBytes(payload, "tool_choice"); tc.Exists() && tc.Get("type").String() == "tool" {
		tcName := tc.Get("name").String()
		if alias, exists := rewrite.Forward[tcName]; exists {
			if patched, patchErr := sjson.SetBytes(payload, "tool_choice.name", alias); patchErr == nil {
				payload = patched
			}
		}
	}

	// Rewrite tool_use references in message history.
	msgsResult := gjson.GetBytes(payload, "messages")
	if msgsResult.IsArray() {
		msgsResult.ForEach(func(mKey, mVal gjson.Result) bool {
			mIdx := int(mKey.Num)
			contentResult := mVal.Get("content")
			if !contentResult.IsArray() {
				return true
			}
			contentResult.ForEach(func(cKey, cVal gjson.Result) bool {
				cIdx := int(cKey.Num)
				if cVal.Get("type").String() != "tool_use" {
					return true
				}
				nm := cVal.Get("name").String()
				if nm == "" {
					return true
				}
				if alias, exists := rewrite.Forward[nm]; exists {
					jp := fmt.Sprintf("messages.%d.content.%d.name", mIdx, cIdx)
					if patched, patchErr := sjson.SetBytes(payload, jp, alias); patchErr == nil {
						payload = patched
					}
				}
				return true
			})
			return true
		})
	}

	return applyToolsLastCacheBreakpoint(payload)
}

// applyToolsLastCacheBreakpoint injects a cache_control breakpoint on the last tool
// in the tools array. TTL policy:
//   - Client-set cache_control.ttl is preserved.
//   - Otherwise {"type":"ephemeral","ttl": claude.DefaultCacheControlTTL} is written.
//
// No-op when tools is absent or empty.
func applyToolsLastCacheBreakpoint(payload []byte) []byte {
	toolsResult := gjson.GetBytes(payload, "tools")
	if !toolsResult.IsArray() {
		return payload
	}
	items := toolsResult.Array()
	if len(items) == 0 {
		return payload
	}
	tail := len(items) - 1
	cc := items[tail].Get("cache_control")

	if cc.Exists() && cc.Get("ttl").String() != "" {
		return payload
	}

	if cc.Exists() {
		if patched, patchErr := sjson.SetBytes(payload, fmt.Sprintf("tools.%d.cache_control.ttl", tail), claude.DefaultCacheControlTTL); patchErr == nil {
			payload = patched
		}
		return payload
	}

	fragment := fmt.Sprintf(`{"type":"ephemeral","ttl":%q}`, claude.DefaultCacheControlTTL)
	if patched, patchErr := sjson.SetRawBytes(payload, fmt.Sprintf("tools.%d.cache_control", tail), []byte(fragment)); patchErr == nil {
		payload = patched
	}
	return payload
}

// restoreToolNamesInBytes performs byte-level fake-to-real name restoration on a chunk.
// Processes ReverseOrdered by descending fake-name length to avoid substring collisions,
// then applies static prefix reversal (cc_sess_ -> sessions_, cc_ses_ -> session_).
//
// rw may be nil; static prefix reversal still runs in that case.
func restoreToolNamesInBytes(chunk []byte, rewrite *ToolNameRewrite) []byte {
	if rewrite != nil {
		for _, mapping := range rewrite.ReverseOrdered {
			alias, orig := mapping[0], mapping[1]
			if alias == "" || alias == orig {
				continue
			}
			chunk = replaceAllBytes(chunk, alias, orig)
		}
	}
	for origPrefix, fakePrefix := range staticToolNameRewrites {
		chunk = replaceAllBytes(chunk, fakePrefix, origPrefix)
	}
	return chunk
}

// replaceAllBytes replaces all occurrences of `from` with `to` in data.
func replaceAllBytes(buf []byte, from, to string) []byte {
	if len(buf) == 0 || from == to {
		return buf
	}
	if !strings.Contains(string(buf), from) {
		return buf
	}
	return []byte(strings.ReplaceAll(string(buf), from, to))
}

// toolNameRewriteFromContext retrieves the ToolNameRewrite stored during request processing.
// Returns nil when not found or type assertion fails; callers must handle nil.
func toolNameRewriteFromContext(ctx interface {
	Get(string) (any, bool)
}) *ToolNameRewrite {
	if ctx == nil {
		return nil
	}
	val, exists := ctx.Get(toolNameRewriteKey)
	if !exists || val == nil {
		return nil
	}
	typed, _ := val.(*ToolNameRewrite)
	return typed
}

// reverseToolNamesIfPresent is the unified response-side entry point: extracts the
// mapping from context and performs byte-level fake-to-real restoration. Static prefix
// reversal runs even when no dynamic mapping exists.
func reverseToolNamesIfPresent(ctx interface {
	Get(string) (any, bool)
}, chunk []byte) []byte {
	rewrite := toolNameRewriteFromContext(ctx)
	if rewrite == nil && len(staticToolNameRewrites) == 0 {
		return chunk
	}
	return restoreToolNamesInBytes(chunk, rewrite)
}
