package service

import (
	"context"
	"testing"
)

type patchCacheStub struct {
	SchedulerCache
	patches []patchCall
}

type patchCall struct {
	bucket    SchedulerBucket
	accountID int64
	belongs   bool
	priority  int
}

func (s *patchCacheStub) PatchAccountInSnapshot(_ context.Context, bucket SchedulerBucket, accountID int64, belongs bool, priority int) (bool, error) {
	s.patches = append(s.patches, patchCall{bucket, accountID, belongs, priority})
	return true, nil
}

func (s *patchCacheStub) SetAccount(_ context.Context, _ *Account) error { return nil }

func TestTryPatchBuckets_SinglePlatform(t *testing.T) {
	cache := &patchCacheStub{}
	svc := &SchedulerSnapshotService{cache: cache}

	account := &Account{
		ID:       42,
		Platform: PlatformAnthropic,
		Status:   StatusActive,
		Schedulable: true,
		Priority: 5,
		GroupIDs: []int64{1, 2},
	}

	ok := svc.tryPatchBuckets(context.Background(), account, []int64{1, 2}, nil)
	if !ok {
		t.Fatal("expected patch to succeed")
	}

	// Anthropic has 3 modes (single, forced, mixed) × 2 groups = 6 patches
	if len(cache.patches) != 6 {
		t.Fatalf("expected 6 patches, got %d", len(cache.patches))
	}

	for _, p := range cache.patches {
		if p.accountID != 42 {
			t.Errorf("wrong account ID: %d", p.accountID)
		}
		if !p.belongs {
			t.Error("account should belong to its own groups")
		}
		if p.priority != 5 {
			t.Errorf("wrong priority: %d", p.priority)
		}
	}
}

func TestTryPatchBuckets_UnschedulableAccount(t *testing.T) {
	cache := &patchCacheStub{}
	svc := &SchedulerSnapshotService{cache: cache}

	account := &Account{
		ID:          42,
		Platform:    PlatformAnthropic,
		Status:      "error",
		Schedulable: true,
		GroupIDs:    []int64{1},
	}

	ok := svc.tryPatchBuckets(context.Background(), account, []int64{1}, nil)
	if !ok {
		t.Fatal("expected patch to succeed")
	}

	for _, p := range cache.patches {
		if p.belongs {
			t.Error("error-status account should not belong to any bucket")
		}
	}
}

func TestTryPatchBuckets_AccountNotInGroup(t *testing.T) {
	cache := &patchCacheStub{}
	svc := &SchedulerSnapshotService{cache: cache}

	account := &Account{
		ID:          42,
		Platform:    PlatformAnthropic,
		Status:      StatusActive,
		Schedulable: true,
		GroupIDs:    []int64{1},
	}

	ok := svc.tryPatchBuckets(context.Background(), account, []int64{1, 99}, nil)
	if !ok {
		t.Fatal("expected patch to succeed")
	}

	var belongsCount, notBelongsCount int
	for _, p := range cache.patches {
		if p.bucket.GroupID == 1 && p.belongs {
			belongsCount++
		}
		if p.bucket.GroupID == 99 && !p.belongs {
			notBelongsCount++
		}
	}
	if belongsCount == 0 {
		t.Error("account should belong to group 1 buckets")
	}
	if notBelongsCount == 0 {
		t.Error("account should NOT belong to group 99 buckets")
	}
}
