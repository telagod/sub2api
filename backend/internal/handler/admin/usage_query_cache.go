package admin

import (
	"context"
	"time"

	"github.com/telagod/subme/internal/pkg/usagestats"
)

var usageStatsCache = newSnapshotCache(30 * time.Second)

type usageStatsCacheKeyData struct {
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	UserID      int64  `json:"user_id"`
	APIKeyID    int64  `json:"api_key_id"`
	AccountID   int64  `json:"account_id"`
	GroupID     int64  `json:"group_id"`
	Model       string `json:"model"`
	BillingMode string `json:"billing_mode"`
	RequestType *int16 `json:"request_type"`
	Stream      *bool  `json:"stream"`
	BillingType *int8  `json:"billing_type"`
}

func usageStatsCacheKey(f usagestats.UsageLogFilters) string {
	startStr := ""
	if f.StartTime != nil {
		startStr = f.StartTime.UTC().Format(time.RFC3339)
	}
	endStr := ""
	if f.EndTime != nil {
		endStr = f.EndTime.UTC().Format(time.RFC3339)
	}
	return mustMarshalDashboardCacheKey(usageStatsCacheKeyData{
		StartTime:   startStr,
		EndTime:     endStr,
		UserID:      f.UserID,
		APIKeyID:    f.APIKeyID,
		AccountID:   f.AccountID,
		GroupID:     f.GroupID,
		Model:       f.Model,
		BillingMode: f.BillingMode,
		RequestType: f.RequestType,
		Stream:      f.Stream,
		BillingType: f.BillingType,
	})
}

// getStatsCached returns cached usage stats or loads from the service on miss.
func (h *UsageHandler) getStatsCached(ctx context.Context, f usagestats.UsageLogFilters) (*usagestats.UsageStats, bool, error) {
	cacheKey := usageStatsCacheKey(f)
	cached, wasHit, loadErr := usageStatsCache.GetOrLoad(cacheKey, func() (any, error) {
		return h.usageService.GetStatsWithFilters(ctx, f)
	})
	if loadErr != nil {
		return nil, wasHit, loadErr
	}
	result, castErr := snapshotPayloadAs[*usagestats.UsageStats](cached.Payload)
	return result, wasHit, castErr
}
