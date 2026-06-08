package service

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
)

type groupSummaryRepoStub struct {
	UsageLogRepository
	calls  int
	result []usagestats.GroupUsageSummary
}

func (r *groupSummaryRepoStub) GetAllGroupUsageSummary(_ context.Context, _ time.Time) ([]usagestats.GroupUsageSummary, error) {
	r.calls++
	return r.result, nil
}

func TestGetGroupUsageSummary_CachesResult(t *testing.T) {
	// Reset global cache
	groupUsageSummaryCache = atomic.Value{}

	repo := &groupSummaryRepoStub{
		result: []usagestats.GroupUsageSummary{
			{GroupID: 1, TotalCost: 100, TodayCost: 10},
			{GroupID: 2, TotalCost: 200, TodayCost: 20},
		},
	}
	svc := &DashboardService{usageRepo: repo}
	ctx := context.Background()
	todayStart := time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC)

	// First call should hit repo
	results, err := svc.GetGroupUsageSummary(ctx, todayStart)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if repo.calls != 1 {
		t.Fatalf("expected 1 repo call, got %d", repo.calls)
	}

	// Second call should use cache
	results2, err := svc.GetGroupUsageSummary(ctx, todayStart)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results2) != 2 {
		t.Fatalf("expected 2 cached results, got %d", len(results2))
	}
	if repo.calls != 1 {
		t.Fatalf("cache miss: expected 1 repo call, got %d", repo.calls)
	}

	// Different todayStart should bypass cache
	differentDay := todayStart.Add(24 * time.Hour)
	_, _ = svc.GetGroupUsageSummary(ctx, differentDay)
	if repo.calls != 2 {
		t.Fatalf("expected 2 repo calls for different day, got %d", repo.calls)
	}
}
