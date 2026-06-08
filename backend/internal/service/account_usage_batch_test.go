package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
)

type batchRepoStub struct {
	UsageLogRepository
	batchResult map[int64]*usagestats.AccountStats
	batchCalled bool
}

func (s *batchRepoStub) GetAccountWindowStatsBatch(_ context.Context, ids []int64, _ time.Time) (map[int64]*usagestats.AccountStats, error) {
	s.batchCalled = true
	out := make(map[int64]*usagestats.AccountStats, len(ids))
	for _, id := range ids {
		if v, ok := s.batchResult[id]; ok {
			out[id] = v
		}
	}
	return out, nil
}

func (s *batchRepoStub) GetAccountWindowStats(_ context.Context, id int64, _ time.Time) (*usagestats.AccountStats, error) {
	if v, ok := s.batchResult[id]; ok {
		return v, nil
	}
	return &usagestats.AccountStats{}, nil
}

func TestGetAccountWindowStatsBatch_UsesBatchPath(t *testing.T) {
	repo := &batchRepoStub{
		batchResult: map[int64]*usagestats.AccountStats{
			1: {Requests: 10, StandardCost: 1.5},
			2: {Requests: 20, StandardCost: 3.0},
		},
	}
	svc := &AccountUsageService{usageLogRepo: repo}

	result, err := svc.GetAccountWindowStatsBatch(context.Background(), []int64{1, 2, 3}, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.batchCalled {
		t.Fatal("expected batch path to be used")
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
	if result[1].Requests != 10 || result[1].StandardCost != 1.5 {
		t.Errorf("account 1: got requests=%d cost=%.2f", result[1].Requests, result[1].StandardCost)
	}
	if result[2].Requests != 20 || result[2].StandardCost != 3.0 {
		t.Errorf("account 2: got requests=%d cost=%.2f", result[2].Requests, result[2].StandardCost)
	}
	if _, ok := result[3]; ok {
		t.Error("account 3 should not be in results (no data)")
	}
}

func TestGetAccountWindowStatsBatch_EmptyInput(t *testing.T) {
	repo := &batchRepoStub{batchResult: map[int64]*usagestats.AccountStats{}}
	svc := &AccountUsageService{usageLogRepo: repo}

	result, err := svc.GetAccountWindowStatsBatch(context.Background(), []int64{}, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %d", len(result))
	}
}
