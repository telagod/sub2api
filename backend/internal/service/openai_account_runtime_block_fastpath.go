package service

import (
	"context"
	"net/http"
	"time"
)

const (
	openAIAccountStateUpdateTimeout       = 5 * time.Second
	openAIOAuth429FallbackCooldown        = 5 * time.Second
	openAIStopSchedulingBridgeCooldown    = 2 * time.Minute
	openAIOAuth429StormWindow             = 10 * time.Second
	openAIOAuth429StormThreshold          = 20
	openAIOAuth429StormMaxAccountSwitches = 1
)

func openAIAccountStateContext(parentCtx context.Context) (context.Context, context.CancelFunc) {
	base := context.Background()
	if parentCtx != nil {
		base = context.WithoutCancel(parentCtx)
	}
	return context.WithTimeout(base, openAIAccountStateUpdateTimeout)
}

func isOpenAIOAuthAccount(acct *Account) bool {
	return acct != nil && acct.Platform == PlatformOpenAI && acct.Type == AccountTypeOAuth
}

func isOpenAIAccount(acct *Account) bool {
	return acct != nil && acct.Platform == PlatformOpenAI
}

func (s *OpenAIGatewayService) handleOpenAIAccountUpstreamError(reqCtx context.Context, acct *Account, statusCode int, hdrs http.Header, respBody []byte, requestedModel ...string) bool {
	stateCtx, cancel := openAIAccountStateContext(reqCtx)
	defer cancel()

	if checkOpenAIImageRateLimitError(statusCode, respBody) {
		if s != nil && s.rateLimitService != nil {
			_ = s.rateLimitService.HandleOpenAIImageRateLimit(stateCtx, acct, statusCode, hdrs, respBody)
		}
		return false
	}

	if statusCode == http.StatusTooManyRequests {
		s.markOpenAIOAuth429RateLimited(stateCtx, acct, hdrs, respBody)
	}
	if s == nil || acct == nil || s.rateLimitService == nil {
		return false
	}
	if len(requestedModel) > 0 && s.rateLimitService.HandleUpstreamModelNotFound(stateCtx, acct, requestedModel[0], statusCode, respBody) {
		return true
	}
	needDisable := s.rateLimitService.HandleUpstreamError(stateCtx, acct, statusCode, hdrs, respBody)
	if needDisable {
		s.BlockAccountScheduling(acct, time.Time{}, "upstream_disable")
	}
	return needDisable
}

func (s *OpenAIGatewayService) markOpenAIOAuth429RateLimited(reqCtx context.Context, acct *Account, hdrs http.Header, respBody []byte) {
	if s == nil || !isOpenAIOAuthAccount(acct) {
		return
	}
	s.recordOpenAIOAuth429()

	cooldown := time.Now().Add(openAIOAuth429FallbackCooldown)
	if s.rateLimitService != nil {
		if resetAt := s.rateLimitService.calcOpenAI429ResetTime(hdrs); resetAt != nil && resetAt.After(time.Now()) {
			cooldown = *resetAt
		} else if resetUnix := parseOpenAIRateLimitResetTime(respBody); resetUnix != nil {
			if ts := time.Unix(*resetUnix, 0); ts.After(time.Now()) {
				cooldown = ts
			}
		} else if dur, found := s.rateLimitService.get429FallbackCooldown(reqCtx, acct); found && dur > 0 {
			cooldown = time.Now().Add(dur)
		}
	}
	s.BlockAccountScheduling(acct, cooldown, "429")
}

// BlockAccountScheduling prevents the given account from being scheduled
// until the specified deadline. Uses a CAS loop for concurrent safety.
func (s *OpenAIGatewayService) BlockAccountScheduling(acct *Account, until time.Time, reason string) {
	if s == nil || !isOpenAIAccount(acct) {
		return
	}
	nowTs := time.Now()
	deadline := until
	if deadline.IsZero() || !deadline.After(nowTs) {
		deadline = nowTs.Add(openAIStopSchedulingBridgeCooldown)
	}

	for {
		existing, loaded := s.openaiAccountRuntimeBlockUntil.Load(acct.ID)
		if !loaded {
			actual, raced := s.openaiAccountRuntimeBlockUntil.LoadOrStore(acct.ID, deadline)
			if !raced {
				return
			}
			existing = actual
		}

		existingTs, valid := existing.(time.Time)
		if !valid || existingTs.IsZero() {
			if s.openaiAccountRuntimeBlockUntil.CompareAndSwap(acct.ID, existing, deadline) {
				return
			}
			continue
		}
		if existingTs.After(deadline) {
			return
		}
		if s.openaiAccountRuntimeBlockUntil.CompareAndSwap(acct.ID, existing, deadline) {
			return
		}
	}
}

// ClearAccountSchedulingBlock removes the runtime block for the given account.
func (s *OpenAIGatewayService) ClearAccountSchedulingBlock(acctID int64) {
	if s == nil || acctID <= 0 {
		return
	}
	s.openaiAccountRuntimeBlockUntil.Delete(acctID)
}

func (s *OpenAIGatewayService) isOpenAIAccountRuntimeBlocked(acct *Account) bool {
	if s == nil || !isOpenAIAccount(acct) {
		return false
	}
	val, loaded := s.openaiAccountRuntimeBlockUntil.Load(acct.ID)
	if !loaded {
		return false
	}
	deadline, valid := val.(time.Time)
	if !valid || deadline.IsZero() {
		s.openaiAccountRuntimeBlockUntil.Delete(acct.ID)
		return false
	}
	if time.Now().Before(deadline) {
		return true
	}
	s.openaiAccountRuntimeBlockUntil.Delete(acct.ID)
	return false
}

func (s *OpenAIGatewayService) recordOpenAIOAuth429() {
	if s == nil {
		return
	}
	nowTs := time.Now()
	winStart := s.openaiOAuth429WindowStartUnixNano.Load()
	if winStart == 0 || nowTs.Sub(time.Unix(0, winStart)) >= openAIOAuth429StormWindow {
		if s.openaiOAuth429WindowStartUnixNano.CompareAndSwap(winStart, nowTs.UnixNano()) {
			s.openaiOAuth429WindowCount.Store(1)
			return
		}
	}
	s.openaiOAuth429WindowCount.Add(1)
}

func (s *OpenAIGatewayService) isOpenAIOAuth429Storm() bool {
	if s == nil {
		return false
	}
	winStart := s.openaiOAuth429WindowStartUnixNano.Load()
	if winStart == 0 || time.Since(time.Unix(0, winStart)) >= openAIOAuth429StormWindow {
		return false
	}
	return s.openaiOAuth429WindowCount.Load() >= openAIOAuth429StormThreshold
}

// ShouldStopOpenAIOAuth429Failover returns true when the 429 storm detector
// advises against further account switches for the current request.
func (s *OpenAIGatewayService) ShouldStopOpenAIOAuth429Failover(acct *Account, statusCode int, switchAttempts int) bool {
	if statusCode != http.StatusTooManyRequests || switchAttempts < openAIOAuth429StormMaxAccountSwitches {
		return false
	}
	if !isOpenAIOAuthAccount(acct) {
		return false
	}
	return s.isOpenAIOAuth429Storm()
}
