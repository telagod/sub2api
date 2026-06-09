package service

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/telagod/subme/internal/pkg/logger"
	"go.uber.org/zap"
)

// openAITransportErrorTempUnschedDuration controls how long an account stays
// unschedulable after a durable transport fault.
const openAITransportErrorTempUnschedDuration = 10 * time.Minute

// openAITransportFailoverBody is the standard OpenAI-format error body
// attached to the failover error for transport-level failures.
var openAITransportFailoverBody = []byte(`{"error":{"type":"upstream_error","message":"Upstream request failed"}}`)

// openAITransportErrorClass describes the reaction to a transport-level
// upstream failure where no HTTP response was received.
type openAITransportErrorClass struct {
	// Persistent indicates that retrying the same proxy/account is futile:
	// expired proxy credentials, dead endpoint, DNS failure, etc.
	Persistent bool
}

// openAIPersistentTransportErrorMarkers are substrings (matched
// case-insensitively) that signal a durable proxy/network fault.
// Only specific failure reasons are matched to avoid misclassifying
// transient issues (e.g. a proxy timeout) as permanent.
var openAIPersistentTransportErrorMarkers = []string{
	"authentication failed",
	"proxy authentication required",
	"connection refused",
	"no route to host",
	"network is unreachable",
	"no such host",
}

// classifyOpenAITransportError determines whether a transport error is
// permanent (Persistent -- evict account + alert) or transient (fail over
// but keep account schedulable).
func classifyOpenAITransportError(opErr error) openAITransportErrorClass {
	if opErr == nil {
		return openAITransportErrorClass{}
	}

	// Typed error checks (preferred: portable and unambiguous)
	if errors.Is(opErr, syscall.ECONNREFUSED) ||
		errors.Is(opErr, syscall.EHOSTUNREACH) ||
		errors.Is(opErr, syscall.ENETUNREACH) {
		return openAITransportErrorClass{Persistent: true}
	}
	var dnsFailure *net.DNSError
	if errors.As(opErr, &dnsFailure) && dnsFailure.IsNotFound {
		return openAITransportErrorClass{Persistent: true}
	}

	// String-marker fallback for errors lacking typed forms
	lowered := strings.ToLower(opErr.Error())
	for _, marker := range openAIPersistentTransportErrorMarkers {
		if strings.Contains(lowered, marker) {
			return openAITransportErrorClass{Persistent: true}
		}
	}
	return openAITransportErrorClass{}
}

// handleOpenAIUpstreamTransportError processes a transport-level upstream
// failure (proxy/DNS/TCP/TLS error, no HTTP status received). It records
// the fault in the ops error log, optionally evicts the account for durable
// faults, and returns a failover error for non-canceled contexts.
//
// The method does NOT write to the HTTP response; the caller owns response
// writing (failover or protocol-correct error after failover exhaustion).
func (s *OpenAIGatewayService) handleOpenAIUpstreamTransportError(reqCtx context.Context, gc *gin.Context, acct *Account, opErr error, passthrough bool) error {
	sanitized := sanitizeUpstreamErrorMessage(opErr.Error())
	setOpsUpstreamError(gc, 0, sanitized, "")
	appendOpsUpstreamError(gc, OpsUpstreamErrorEvent{
		Platform:           acct.Platform,
		AccountID:          acct.ID,
		AccountName:        acct.Name,
		UpstreamStatusCode: 0,
		Passthrough:        passthrough,
		Kind:               "request_error",
		Message:            sanitized,
	})

	// Client already gone: no failover, no eviction.
	if errors.Is(opErr, context.Canceled) {
		return opErr
	}

	if classifyOpenAITransportError(opErr).Persistent {
		s.tempUnscheduleOpenAITransportError(reqCtx, acct, sanitized)
	}

	return &UpstreamFailoverError{
		StatusCode:   http.StatusBadGateway,
		ResponseBody: openAITransportFailoverBody,
	}
}

// tempUnscheduleOpenAITransportError marks an account temporarily unschedulable
// after a durable transport failure, both in-memory (immediate scheduler effect)
// and in the DB (survives restart).
func (s *OpenAIGatewayService) tempUnscheduleOpenAITransportError(reqCtx context.Context, acct *Account, sanitized string) {
	if s == nil || acct == nil {
		return
	}
	deadline := time.Now().Add(openAITransportErrorTempUnschedDuration)
	evictReason := "upstream transport error (proxy/network): " + sanitized

	// Immediate in-memory block so the scheduler stops selecting this account
	// right away, even before the DB write propagates.
	s.BlockAccountScheduling(acct, deadline, "transport_error")

	if s.accountRepo == nil {
		// No DB: in-memory only. Emit a distinct log event so operators know
		// the block will not survive a process restart.
		logger.L().With(zap.String("component", "service.openai_gateway")).Warn(
			"openai.account_temp_unscheduled_transport_memory_only",
			zap.Int64("account_id", acct.ID),
			zap.String("account_name", acct.Name),
			zap.String("platform", acct.Platform),
			zap.Time("until", deadline),
			zap.String("reason", evictReason),
		)
		return
	}

	bgCtx, cancel := context.WithTimeout(context.WithoutCancel(reqCtx), openAIAccountStateUpdateTimeout)
	defer cancel()
	if dbErr := s.accountRepo.SetTempUnschedulable(bgCtx, acct.ID, deadline, evictReason); dbErr != nil {
		logger.L().With(zap.String("component", "service.openai_gateway")).Warn(
			"openai.account_temp_unscheduled_transport_failed",
			zap.Int64("account_id", acct.ID),
			zap.Error(dbErr),
		)
		return
	}

	// DB write succeeded: both in-memory and persisted.
	logger.L().With(zap.String("component", "service.openai_gateway")).Warn(
		"openai.account_temp_unscheduled_transport",
		zap.Int64("account_id", acct.ID),
		zap.String("account_name", acct.Name),
		zap.String("platform", acct.Platform),
		zap.Time("until", deadline),
		zap.String("reason", evictReason),
	)
}
