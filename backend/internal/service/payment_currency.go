package service

import (
	"strings"

	dbent "github.com/telagod/subme/ent"
	"github.com/telagod/subme/internal/payment"
)

func paymentProviderConfigCurrency(provider string, cfg map[string]string) string {
	trimmed := strings.TrimSpace(provider)
	if trimmed == payment.TypeStripe || trimmed == payment.TypeAirwallex {
		if cur, e := payment.NormalizePaymentCurrency(cfg["currency"]); e == nil {
			return cur
		}
	}
	return payment.DefaultPaymentCurrency
}

func PaymentOrderCurrency(order *dbent.PaymentOrder) string {
	snap := psOrderProviderSnapshot(order)
	if snap != nil {
		if cur, e := payment.NormalizePaymentCurrency(snap.Currency); e == nil {
			return cur
		}
	}
	return payment.DefaultPaymentCurrency
}
