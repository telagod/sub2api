package service

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/telagod/subme/internal/pkg/apicompat"
	"github.com/tidwall/gjson"
)

type openAICompatSessionResponseBinding struct {
	ResponseID           string
	TurnState            string
	ContinuationDisabled bool
	ExpiresAt            time.Time
}

// openAICompatContinuationEnabled checks whether session-level continuation
// is available for the given account and model combination.
func openAICompatContinuationEnabled(acct *Account, modelName string) bool {
	if acct == nil || acct.Type != AccountTypeAPIKey {
		return false
	}
	return shouldAutoInjectPromptCacheKeyForCompat(modelName)
}

// trimAnthropicCompatResponsesInputToLatestTurn strips earlier turns from
// the input slice, keeping only the most recent turn boundary.
func trimAnthropicCompatResponsesInputToLatestTurn(req *apicompat.ResponsesRequest) {
	if req == nil || len(req.Input) == 0 {
		return
	}

	var parsed []apicompat.ResponsesInputItem
	if unmarshalErr := json.Unmarshal(req.Input, &parsed); unmarshalErr != nil || len(parsed) == 0 {
		return
	}

	boundary := latestAnthropicCompatResponsesInputTurnStart(parsed)
	kept := append([]apicompat.ResponsesInputItem(nil), parsed[boundary:]...)
	if len(kept) == len(parsed) {
		return
	}
	if encoded, encErr := json.Marshal(kept); encErr == nil {
		req.Input = encoded
	}
}

// latestAnthropicCompatResponsesInputTurnStart locates the start index of
// the most recent logical turn in the input item list.
func latestAnthropicCompatResponsesInputTurnStart(elems []apicompat.ResponsesInputItem) int {
	if len(elems) == 0 {
		return 0
	}

	idx := len(elems) - 1
	tail := elems[idx]
	switch {
	case tail.Type == "function_call_output":
		for idx > 0 && elems[idx-1].Type == "function_call_output" {
			idx--
		}
	case tail.Type == "message" && tail.Role == "user":
		for idx > 0 && elems[idx-1].Type == "function_call_output" {
			idx--
		}
	default:
		return idx
	}

	return expandAnthropicCompatResponsesInputToolCallStart(elems, idx)
}

// expandAnthropicCompatResponsesInputToolCallStart walks backwards to include
// any function_call items whose call_id is referenced by downstream
// function_call_output items.
func expandAnthropicCompatResponsesInputToolCallStart(elems []apicompat.ResponsesInputItem, pos int) int {
	if pos < 0 || pos >= len(elems) {
		return pos
	}

	required := make(map[string]struct{})
	for i := pos; i < len(elems); i++ {
		if elems[i].Type != "function_call_output" {
			continue
		}
		cid := strings.TrimSpace(elems[i].CallID)
		if cid != "" {
			required[cid] = struct{}{}
		}
	}
	if len(required) == 0 {
		return pos
	}

	expanded := pos
	for i := pos - 1; i >= 0 && len(required) > 0; i-- {
		if elems[i].Type != "function_call" {
			continue
		}
		cid := strings.TrimSpace(elems[i].CallID)
		if _, found := required[cid]; !found {
			continue
		}
		delete(required, cid)
		expanded = i
	}
	return expanded
}

// isOpenAICompatPreviousResponseNotFound returns true when the upstream
// reports that the referenced previous_response_id does not exist.
func isOpenAICompatPreviousResponseNotFound(statusCode int, errMsg string, errBody []byte) bool {
	if statusCode != http.StatusBadRequest && statusCode != http.StatusNotFound {
		return false
	}
	probe := func(text string) bool {
		lo := strings.ToLower(strings.TrimSpace(text))
		return strings.Contains(lo, "previous_response_not_found") ||
			(strings.Contains(lo, "previous response") && strings.Contains(lo, "not found")) ||
			(strings.Contains(lo, "unsupported parameter") && strings.Contains(lo, "previous_response_id"))
	}
	if probe(errMsg) || probe(string(errBody)) {
		return true
	}
	return probe(gjson.GetBytes(errBody, "error.code").String()) ||
		probe(gjson.GetBytes(errBody, "error.message").String())
}

// isOpenAICompatPreviousResponseUnsupported returns true when the upstream
// indicates that previous_response_id is not a supported parameter at all.
func isOpenAICompatPreviousResponseUnsupported(statusCode int, errMsg string, errBody []byte) bool {
	if statusCode != http.StatusBadRequest {
		return false
	}
	probe := func(text string) bool {
		lo := strings.ToLower(strings.TrimSpace(text))
		if !strings.Contains(lo, "previous_response_id") {
			return false
		}
		return strings.Contains(lo, "unsupported parameter") ||
			strings.Contains(lo, "only supported on responses websocket") ||
			strings.Contains(lo, "not supported")
	}
	if probe(errMsg) || probe(string(errBody)) {
		return true
	}
	return probe(gjson.GetBytes(errBody, "error.code").String()) ||
		probe(gjson.GetBytes(errBody, "error.message").String())
}

func openAICompatSessionResponseKey(gc *gin.Context, acct *Account, cacheKey string) string {
	ck := strings.TrimSpace(cacheKey)
	if acct == nil || ck == "" {
		return ""
	}
	keyID := int64(0)
	if gc != nil {
		keyID = getAPIKeyIDFromContext(gc)
	}
	return strings.Join([]string{
		strconv.FormatInt(acct.ID, 10),
		strconv.FormatInt(keyID, 10),
		ck,
	}, "\x00")
}

func (s *OpenAIGatewayService) getOpenAICompatSessionResponseID(_ context.Context, gc *gin.Context, acct *Account, cacheKey string) string {
	if s == nil {
		return ""
	}
	sessionKey := openAICompatSessionResponseKey(gc, acct, cacheKey)
	if sessionKey == "" {
		return ""
	}
	val, loaded := s.openaiCompatSessionResponses.Load(sessionKey)
	if !loaded {
		return ""
	}
	entry, valid := val.(openAICompatSessionResponseBinding)
	if !valid {
		s.openaiCompatSessionResponses.Delete(sessionKey)
		return ""
	}
	if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		s.openaiCompatSessionResponses.Delete(sessionKey)
		return ""
	}
	if entry.ContinuationDisabled {
		return ""
	}
	rid := strings.TrimSpace(entry.ResponseID)
	if rid == "" {
		s.openaiCompatSessionResponses.Delete(sessionKey)
		return ""
	}
	return rid
}

func (s *OpenAIGatewayService) bindOpenAICompatSessionResponseID(_ context.Context, gc *gin.Context, acct *Account, cacheKey, respID string) {
	if s == nil {
		return
	}
	sessionKey := openAICompatSessionResponseKey(gc, acct, cacheKey)
	rid := strings.TrimSpace(respID)
	if sessionKey == "" || rid == "" {
		return
	}
	entry := openAICompatSessionResponseBinding{
		ResponseID: rid,
		ExpiresAt:  time.Now().Add(s.openAIWSResponseStickyTTL()),
	}
	if val, loaded := s.openaiCompatSessionResponses.Load(sessionKey); loaded {
		if prev, valid := val.(openAICompatSessionResponseBinding); valid {
			if prev.ContinuationDisabled {
				prev.ResponseID = ""
				prev.ExpiresAt = time.Now().Add(s.openAIWSResponseStickyTTL())
				s.openaiCompatSessionResponses.Store(sessionKey, prev)
				return
			}
			entry.TurnState = prev.TurnState
		}
	}
	s.openaiCompatSessionResponses.Store(sessionKey, entry)
}

func (s *OpenAIGatewayService) deleteOpenAICompatSessionResponseID(_ context.Context, gc *gin.Context, acct *Account, cacheKey string) {
	if s == nil {
		return
	}
	sessionKey := openAICompatSessionResponseKey(gc, acct, cacheKey)
	if sessionKey == "" {
		return
	}
	val, loaded := s.openaiCompatSessionResponses.Load(sessionKey)
	if !loaded {
		return
	}
	entry, valid := val.(openAICompatSessionResponseBinding)
	if !valid {
		s.openaiCompatSessionResponses.Delete(sessionKey)
		return
	}
	entry.ResponseID = ""
	if strings.TrimSpace(entry.TurnState) == "" && !entry.ContinuationDisabled {
		s.openaiCompatSessionResponses.Delete(sessionKey)
		return
	}
	entry.ExpiresAt = time.Now().Add(s.openAIWSResponseStickyTTL())
	s.openaiCompatSessionResponses.Store(sessionKey, entry)
}

func (s *OpenAIGatewayService) disableOpenAICompatSessionContinuation(_ context.Context, gc *gin.Context, acct *Account, cacheKey string) {
	if s == nil {
		return
	}
	sessionKey := openAICompatSessionResponseKey(gc, acct, cacheKey)
	if sessionKey == "" {
		return
	}
	entry := openAICompatSessionResponseBinding{
		ContinuationDisabled: true,
		ExpiresAt:            time.Now().Add(s.openAIWSResponseStickyTTL()),
	}
	if val, loaded := s.openaiCompatSessionResponses.Load(sessionKey); loaded {
		if prev, valid := val.(openAICompatSessionResponseBinding); valid {
			entry.TurnState = prev.TurnState
		}
	}
	s.openaiCompatSessionResponses.Store(sessionKey, entry)
}

func (s *OpenAIGatewayService) isOpenAICompatSessionContinuationDisabled(_ context.Context, gc *gin.Context, acct *Account, cacheKey string) bool {
	if s == nil {
		return false
	}
	sessionKey := openAICompatSessionResponseKey(gc, acct, cacheKey)
	if sessionKey == "" {
		return false
	}
	val, loaded := s.openaiCompatSessionResponses.Load(sessionKey)
	if !loaded {
		return false
	}
	entry, valid := val.(openAICompatSessionResponseBinding)
	if !valid {
		s.openaiCompatSessionResponses.Delete(sessionKey)
		return false
	}
	if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		s.openaiCompatSessionResponses.Delete(sessionKey)
		return false
	}
	return entry.ContinuationDisabled
}

func (s *OpenAIGatewayService) getOpenAICompatSessionTurnState(_ context.Context, gc *gin.Context, acct *Account, cacheKey string) string {
	if s == nil {
		return ""
	}
	sessionKey := openAICompatSessionResponseKey(gc, acct, cacheKey)
	if sessionKey == "" {
		return ""
	}
	val, loaded := s.openaiCompatSessionResponses.Load(sessionKey)
	if !loaded {
		return ""
	}
	entry, valid := val.(openAICompatSessionResponseBinding)
	if !valid || strings.TrimSpace(entry.TurnState) == "" {
		return ""
	}
	if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		s.openaiCompatSessionResponses.Delete(sessionKey)
		return ""
	}
	return strings.TrimSpace(entry.TurnState)
}

func (s *OpenAIGatewayService) bindOpenAICompatSessionTurnState(_ context.Context, gc *gin.Context, acct *Account, cacheKey, state string) {
	if s == nil {
		return
	}
	sessionKey := openAICompatSessionResponseKey(gc, acct, cacheKey)
	trimmedState := strings.TrimSpace(state)
	if sessionKey == "" || trimmedState == "" {
		return
	}
	entry := openAICompatSessionResponseBinding{
		TurnState: trimmedState,
		ExpiresAt: time.Now().Add(s.openAIWSResponseStickyTTL()),
	}
	if val, loaded := s.openaiCompatSessionResponses.Load(sessionKey); loaded {
		if prev, valid := val.(openAICompatSessionResponseBinding); valid {
			entry.ResponseID = prev.ResponseID
			entry.ContinuationDisabled = prev.ContinuationDisabled
		}
	}
	s.openaiCompatSessionResponses.Store(sessionKey, entry)
}
