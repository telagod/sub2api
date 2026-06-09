package service

import (
	"context"
	"fmt"
	"strings"

	dbent "github.com/telagod/subme/ent"
	infraerrors "github.com/telagod/subme/internal/pkg/errors"
)

func (s *PaymentService) GetPublicOrderByResumeToken(ctx context.Context, token string) (*dbent.PaymentOrder, error) {
	parsed, parseErr := s.paymentResume().ParseToken(strings.TrimSpace(token))
	if parseErr != nil {
		return nil, parseErr
	}

	record, fetchErr := s.entClient.PaymentOrder.Get(ctx, parsed.OrderID)
	if fetchErr != nil {
		if dbent.IsNotFound(fetchErr) {
			return nil, infraerrors.NotFound("NOT_FOUND", "payment order does not exist")
		}
		return nil, fmt.Errorf("fetching order via resume token: %w", fetchErr)
	}

	if parsed.UserID > 0 && record.UserID != parsed.UserID {
		return nil, resumeTokenMismatchErr()
	}

	providerSnap := psOrderProviderSnapshot(record)
	instanceID := strings.TrimSpace(psStringValueV2(record.ProviderInstanceID))
	provKey := strings.TrimSpace(psStringValueV2(record.ProviderKey))

	if providerSnap != nil {
		if providerSnap.ProviderInstanceID != "" {
			instanceID = providerSnap.ProviderInstanceID
		}
		if providerSnap.ProviderKey != "" {
			provKey = providerSnap.ProviderKey
		}
	}

	if parsed.ProviderInstanceID != "" && instanceID != parsed.ProviderInstanceID {
		return nil, resumeTokenMismatchErr()
	}
	if parsed.ProviderKey != "" && !strings.EqualFold(provKey, parsed.ProviderKey) {
		return nil, resumeTokenMismatchErr()
	}
	if parsed.PaymentType != "" && NormalizeVisibleMethod(record.PaymentType) != NormalizeVisibleMethod(parsed.PaymentType) {
		return nil, resumeTokenMismatchErr()
	}

	if record.Status == OrderStatusPending || record.Status == OrderStatusExpired {
		outcome := s.reconcilePaid(ctx, record)
		if outcome == checkPaidResultAlreadyPaid {
			record, fetchErr = s.entClient.PaymentOrder.Get(ctx, record.ID)
			if fetchErr != nil {
				return nil, fmt.Errorf("reloading order after reconciliation: %w", fetchErr)
			}
		}
	}

	return record, nil
}

func resumeTokenMismatchErr() error {
	return infraerrors.BadRequest("INVALID_RESUME_TOKEN", "token does not correspond to the payment order")
}

func (s *PaymentService) ParseWeChatPaymentResumeToken(token string) (*WeChatPaymentResumeClaims, error) {
	return s.paymentResume().ParseWeChatPaymentResumeToken(strings.TrimSpace(token))
}
