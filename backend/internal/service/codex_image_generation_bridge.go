package service

import "strings"

const featureKeyCodexImageGenerationBridge = "codex_image_generation_bridge"

func boolOverridePtr(val bool) *bool {
	return &val
}

func boolOverrideFromMap(m map[string]any, lookupKeys ...string) *bool {
	if m == nil {
		return nil
	}
	for _, k := range lookupKeys {
		if b, ok := m[k].(bool); ok {
			return boolOverridePtr(b)
		}
	}
	return nil
}

func platformBoolOverride(m map[string]any, field string, plat string) *bool {
	if m == nil {
		return nil
	}
	if b, ok := m[field].(bool); ok {
		return boolOverridePtr(b)
	}
	nested, ok := m[field].(map[string]any)
	if !ok {
		return nil
	}
	plat = strings.TrimSpace(plat)
	if plat == "" {
		return nil
	}
	if b, ok := nested[plat].(bool); ok {
		return boolOverridePtr(b)
	}
	return nil
}

// CodexImageGenerationBridgeOverride returns the channel-level override for Codex
// image_generation bridge injection. Nil means follow the global/account policy.
func (c *Channel) CodexImageGenerationBridgeOverride(platform string) *bool {
	if c == nil {
		return nil
	}
	return platformBoolOverride(c.FeaturesConfig, featureKeyCodexImageGenerationBridge, platform)
}

// CodexImageGenerationBridgeOverride returns the account-level override for Codex
// image_generation bridge injection. Nil means follow the channel/global policy.
func (a *Account) CodexImageGenerationBridgeOverride() *bool {
	if a == nil || a.Platform != PlatformOpenAI || a.Extra == nil {
		return nil
	}
	if result := boolOverrideFromMap(a.Extra, featureKeyCodexImageGenerationBridge, "codex_image_generation_bridge_enabled"); result != nil {
		return result
	}
	providerCfg, _ := a.Extra[PlatformOpenAI].(map[string]any)
	return boolOverrideFromMap(providerCfg, featureKeyCodexImageGenerationBridge, "codex_image_generation_bridge_enabled")
}
