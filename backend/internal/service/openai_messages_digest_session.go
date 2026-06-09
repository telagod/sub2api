package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/telagod/subme/internal/pkg/apicompat"
)

type openAICompatAnthropicDigestBinding struct {
	PromptCacheKey string
	ExpiresAt      time.Time
}

// buildOpenAICompatAnthropicDigestChain constructs a deterministic digest
// chain from the Anthropic request by hashing each system prompt and message
// with a role prefix.
func buildOpenAICompatAnthropicDigestChain(req *apicompat.AnthropicRequest) string {
	if req == nil {
		return ""
	}

	segments := make([]string, 0, len(req.Messages)+1)
	if len(req.System) > 0 && strings.TrimSpace(string(req.System)) != "" && strings.TrimSpace(string(req.System)) != "null" {
		segments = append(segments, "s:"+shortHash(req.System))
	}
	for _, msg := range req.Messages {
		body := msg.Content
		if len(body) == 0 || strings.TrimSpace(string(body)) == "" {
			continue
		}
		tag := "u"
		if strings.TrimSpace(msg.Role) == "assistant" {
			tag = "a"
		}
		segments = append(segments, tag+":"+shortHash(body))
	}
	return strings.Join(segments, "-")
}

func openAICompatAnthropicDigestNamespace(acct *Account, clientKeyID int64) string {
	if acct == nil || acct.ID <= 0 {
		return ""
	}
	return fmt.Sprintf("%d|%d|", acct.ID, clientKeyID)
}

// findOpenAICompatAnthropicDigestPromptCacheKey walks the digest chain from
// longest to shortest, looking for a cached prompt cache key binding.
func (s *OpenAIGatewayService) findOpenAICompatAnthropicDigestPromptCacheKey(acct *Account, clientKeyID int64, digestChain string) (promptCacheKey string, matchedChain string) {
	if s == nil || digestChain == "" {
		return "", ""
	}
	ns := openAICompatAnthropicDigestNamespace(acct, clientKeyID)
	if ns == "" {
		return "", ""
	}
	suffix := digestChain
	for {
		if val, loaded := s.openaiCompatAnthropicDigestSessions.Load(ns + suffix); loaded {
			if entry, valid := val.(openAICompatAnthropicDigestBinding); valid {
				if entry.ExpiresAt.IsZero() || time.Now().Before(entry.ExpiresAt) {
					if ck := strings.TrimSpace(entry.PromptCacheKey); ck != "" {
						return ck, suffix
					}
				}
			}
			s.openaiCompatAnthropicDigestSessions.Delete(ns + suffix)
		}
		splitIdx := strings.LastIndex(suffix, "-")
		if splitIdx < 0 {
			return "", ""
		}
		suffix = suffix[:splitIdx]
	}
}

func (s *OpenAIGatewayService) bindOpenAICompatAnthropicDigestPromptCacheKey(acct *Account, clientKeyID int64, digestChain, cacheKey, prevDigestChain string) {
	if s == nil || digestChain == "" || strings.TrimSpace(cacheKey) == "" {
		return
	}
	ns := openAICompatAnthropicDigestNamespace(acct, clientKeyID)
	if ns == "" {
		return
	}
	entry := openAICompatAnthropicDigestBinding{
		PromptCacheKey: strings.TrimSpace(cacheKey),
		ExpiresAt:      time.Now().Add(s.openAIWSResponseStickyTTL()),
	}
	s.openaiCompatAnthropicDigestSessions.Store(ns+digestChain, entry)
	if prevDigestChain != "" && prevDigestChain != digestChain {
		s.openaiCompatAnthropicDigestSessions.Delete(ns + prevDigestChain)
	}
}

func promptCacheKeyFromAnthropicDigest(digestChain string) string {
	if strings.TrimSpace(digestChain) == "" {
		return ""
	}
	return "anthropic-digest-" + hashSensitiveValueForLog(digestChain)
}

func promptCacheKeyFromAnthropicMetadataSession(req *apicompat.AnthropicRequest) string {
	if req == nil || len(req.Metadata) == 0 {
		return ""
	}
	var meta struct {
		UserID string `json:"user_id"`
	}
	if unmarshalErr := json.Unmarshal(req.Metadata, &meta); unmarshalErr != nil {
		return ""
	}
	parsedID := ParseMetadataUserID(meta.UserID)
	if parsedID == nil || strings.TrimSpace(parsedID.SessionID) == "" {
		return ""
	}
	seed := strings.Join([]string{
		"anthropic-metadata",
		strings.TrimSpace(parsedID.DeviceID),
		strings.TrimSpace(parsedID.AccountUUID),
		strings.TrimSpace(parsedID.SessionID),
	}, "|")
	return "anthropic-metadata-" + hashSensitiveValueForLog(seed)
}

func cloneAnthropicRequestForDigest(req *apicompat.AnthropicRequest) *apicompat.AnthropicRequest {
	if req == nil {
		return nil
	}
	dup := *req
	if len(req.System) > 0 {
		dup.System = append(json.RawMessage(nil), req.System...)
	}
	if len(req.Messages) > 0 {
		dup.Messages = append([]apicompat.AnthropicMessage(nil), req.Messages...)
	}
	return &dup
}
