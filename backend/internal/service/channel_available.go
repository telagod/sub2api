package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// AvailableGroupRef holds summary info for a group associated with a channel in the user view.
//
// The user-facing "Available Channels" page uses this to display: exclusive vs public groups
// (IsExclusive), subscription vs standard (SubscriptionType), and default rate multiplier
// (RateMultiplier). User-specific rate overrides are fetched separately via /groups/rates.
type AvailableGroupRef struct {
	ID               int64
	Name             string
	Platform         string
	SubscriptionType string
	RateMultiplier   float64
	IsExclusive      bool
}

// AvailableChannel represents a channel in the "Available Channels" view:
// basic channel info + associated groups + derived supported model list (no wildcards).
type AvailableChannel struct {
	ID                 int64
	Name               string
	Description        string
	Status             string
	BillingModelSource string
	RestrictModels     bool
	Groups             []AvailableGroupRef
	SupportedModels    []SupportedModel
}

// ListAvailable returns the available channel view: each channel with its associated groups
// and supported model list.
//
// Supported models are computed via (*Channel).SupportedModels() (mapping + pricing union).
// For models without channel-specific pricing, global LiteLLM data is used to synthesize
// display-only pricing so users see default prices instead of "not configured".
//
// Group info is looked up via groupRepo.ListActive and mapped by ID; groups that are
// inactive or deleted are silently omitted.
//
// Precondition: s.groupRepo must be non-nil (guaranteed by wire DI). Nil dereference
// is intentional for fail-fast, preventing silent masking of injection failures.
func (s *ChannelService) ListAvailable(reqCtx context.Context) ([]AvailableChannel, error) {
	allChannels, chErr := s.repo.ListAll(reqCtx)
	if chErr != nil {
		return nil, fmt.Errorf("list channels: %w", chErr)
	}

	activeGroups, grpErr := s.groupRepo.ListActive(reqCtx)
	if grpErr != nil {
		return nil, fmt.Errorf("list active groups: %w", grpErr)
	}
	groupLookup := make(map[int64]AvailableGroupRef, len(activeGroups))
	for idx := 0; idx < len(activeGroups); idx++ {
		grp := activeGroups[idx]
		groupLookup[grp.ID] = AvailableGroupRef{
			ID:               grp.ID,
			Name:             grp.Name,
			Platform:         grp.Platform,
			SubscriptionType: grp.SubscriptionType,
			RateMultiplier:   grp.RateMultiplier,
			IsExclusive:      grp.IsExclusive,
		}
	}

	channelViews := make([]AvailableChannel, 0, len(allChannels))
	for idx := 0; idx < len(allChannels); idx++ {
		ch := &allChannels[idx]
		associatedGroups := make([]AvailableGroupRef, 0, len(ch.GroupIDs))
		for _, gID := range ch.GroupIDs {
			if ref, found := groupLookup[gID]; found {
				associatedGroups = append(associatedGroups, ref)
			}
		}
		sort.SliceStable(associatedGroups, func(a, b int) bool {
			return associatedGroups[a].Name < associatedGroups[b].Name
		})

		ch.normalizeBillingModelSource()

		models := ch.SupportedModels()
		s.fillGlobalPricingFallback(models)

		channelViews = append(channelViews, AvailableChannel{
			ID:                 ch.ID,
			Name:               ch.Name,
			Description:        ch.Description,
			Status:             ch.Status,
			BillingModelSource: ch.BillingModelSource,
			RestrictModels:     ch.RestrictModels,
			Groups:             associatedGroups,
			SupportedModels:    models,
		})
	}

	sort.SliceStable(channelViews, func(a, b int) bool {
		return strings.ToLower(channelViews[a].Name) < strings.ToLower(channelViews[b].Name)
	})
	return channelViews, nil
}

// fillGlobalPricingFallback supplements models missing channel-specific pricing with
// global LiteLLM data. Display-only; does not affect the real billing pipeline.
//
// Trigger conditions:
//  1. Pricing == nil (channel has no pricing entry for this model at all)
//  2. Pricing non-nil but all price fields are empty (admin created entry but left prices blank)
//
// When s.pricingService is nil (test scenarios), the fallback is skipped.
func (s *ChannelService) fillGlobalPricingFallback(models []SupportedModel) {
	if s.pricingService == nil {
		return
	}
	for idx := 0; idx < len(models); idx++ {
		if !pricingNeedsFallback(models[idx].Pricing) {
			continue
		}
		globalPricing := s.pricingService.GetModelPricing(models[idx].Name)
		if globalPricing == nil {
			continue
		}
		models[idx].Pricing = synthesizePricingFromLiteLLM(globalPricing, models[idx].Pricing)
	}
}

// pricingNeedsFallback determines whether a ChannelModelPricing needs global fallback.
// Returns true when all price fields are absent (no flat fields and no priced intervals).
func pricingNeedsFallback(pricing *ChannelModelPricing) bool {
	if pricing == nil {
		return true
	}
	if pricing.InputPrice != nil || pricing.OutputPrice != nil ||
		pricing.CacheWritePrice != nil || pricing.CacheReadPrice != nil ||
		pricing.ImageOutputPrice != nil || pricing.PerRequestPrice != nil {
		return false
	}
	for idx := 0; idx < len(pricing.Intervals); idx++ {
		iv := pricing.Intervals[idx]
		if iv.InputPrice != nil || iv.OutputPrice != nil ||
			iv.CacheWritePrice != nil || iv.CacheReadPrice != nil ||
			iv.PerRequestPrice != nil {
			return false
		}
	}
	return true
}

// synthesizePricingFromLiteLLM converts LiteLLM pricing data into ChannelModelPricing format.
// Display-only.
//
// Billing mode priority:
//  1. Channel's already-selected BillingMode (admin picked image/per_request but left prices blank)
//  2. LiteLLM mode="image_generation" -> image
//  3. Default: token
//
// LiteLLM fields with value 0 are treated as unconfigured and excluded from display.
func synthesizePricingFromLiteLLM(globalPricing *LiteLLMModelPricing, existingPricing *ChannelModelPricing) *ChannelModelPricing {
	if globalPricing == nil {
		return existingPricing
	}

	billingMode := BillingModeToken
	if existingPricing != nil && existingPricing.BillingMode != "" {
		billingMode = existingPricing.BillingMode
	} else if globalPricing.Mode == "image_generation" {
		billingMode = BillingModeImage
	}

	if billingMode == BillingModeImage || billingMode == BillingModePerRequest {
		return &ChannelModelPricing{
			BillingMode:      billingMode,
			PerRequestPrice:  nonZeroPtr(globalPricing.OutputCostPerImage),
			ImageOutputPrice: nonZeroPtr(globalPricing.OutputCostPerImageToken),
			InputPrice:       nonZeroPtr(globalPricing.InputCostPerToken),
			OutputPrice:      nonZeroPtr(globalPricing.OutputCostPerToken),
		}
	}
	return &ChannelModelPricing{
		BillingMode:      billingMode,
		InputPrice:       nonZeroPtr(globalPricing.InputCostPerToken),
		OutputPrice:      nonZeroPtr(globalPricing.OutputCostPerToken),
		CacheWritePrice:  nonZeroPtr(globalPricing.CacheCreationInputTokenCost),
		CacheReadPrice:   nonZeroPtr(globalPricing.CacheReadInputTokenCost),
		ImageOutputPrice: nonZeroPtr(globalPricing.OutputCostPerImageToken),
	}
}

func nonZeroPtr(val float64) *float64 {
	if val == 0 {
		return nil
	}
	return &val
}
