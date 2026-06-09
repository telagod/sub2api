// Package config contains validation for DingTalk connection settings.
//
// Security model (internal_only): the DingTalk "internal app" type guarantees
// that only employees of the owning organization can complete OAuth, so
// ValidateDingTalkConfig only requires app_type=internal without mandating a
// separate InternalCorpID check.
package config

import "errors"

var (
	ErrDingTalkV1AppTypeMismatch = errors.New("dingtalk: internal_only requires app_type=internal")
	ErrDingTalkV4InvalidAppKind  = errors.New("dingtalk: dingtalk_app_kind must be internal_app")
)

func ValidateDingTalkConfig(cfg DingTalkConnectConfig) error {
	if !cfg.Enabled {
		return nil
	}
	if cfg.DingTalkAppKind != "internal_app" {
		return ErrDingTalkV4InvalidAppKind
	}
	if cfg.CorpRestrictionPolicy == "internal_only" && cfg.AppType != "internal" {
		return ErrDingTalkV1AppTypeMismatch
	}
	return nil
}
