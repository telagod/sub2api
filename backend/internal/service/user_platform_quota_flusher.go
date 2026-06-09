package service

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/telagod/subme/internal/config"
	"github.com/telagod/subme/internal/pkg/logger"
)

type quotaDirtyCache interface {
	PopDirtyUserPlatformQuotaKeys(ctx context.Context, n int) ([]UserPlatformQuotaKey, error)
	ReaddDirtyUserPlatformQuotaKeys(ctx context.Context, keys []UserPlatformQuotaKey) error
	BatchGetUserPlatformQuotaCache(ctx context.Context, keys []UserPlatformQuotaKey) ([]*UserPlatformQuotaCacheEntry, error)
}

type quotaSnapshotWriter interface {
	BatchSnapshotUsage(ctx context.Context, snapshots []UserPlatformQuotaSnapshot, now time.Time) error
}

type FlusherMetrics struct {
	FlushSuccessTotal     atomic.Int64
	FlushErrorTotal       atomic.Int64
	FlushBatchSizeTotal   atomic.Int64
	FlushLatencyMsMax     atomic.Int64
	DirtyReaddTotal       atomic.Int64
	DirtyLostTotal        atomic.Int64
	FlushFKViolationTotal atomic.Int64
}

const (
	flusherMaxBatchesPerTick = 16
	maxFlushBatchSize        = 6000
	defaultFlushBatchSize    = 1000
)

type UserPlatformQuotaUsageFlusher struct {
	cache        quotaDirtyCache
	quotaRepo    quotaSnapshotWriter
	timingWheel  *TimingWheelService
	enabled      bool
	interval     time.Duration
	batchSize    int
	flushTimeout time.Duration
	metrics      *FlusherMetrics
	stopped      atomic.Bool
}

func NewUserPlatformQuotaUsageFlusher(cfg *config.Config, cache BillingCache, quotaRepo UserPlatformQuotaRepository, tw *TimingWheelService) *UserPlatformQuotaUsageFlusher {
	batch := cfg.Database.UserPlatformQuotaFlushBatchSize
	if batch <= 0 {
		batch = defaultFlushBatchSize
	}
	if batch > maxFlushBatchSize {
		logger.LegacyPrintf("quota_flusher", "[QuotaFlusher] batch size %d clamped to %d", batch, maxFlushBatchSize)
		batch = maxFlushBatchSize
	}
	iv := time.Duration(cfg.Database.UserPlatformQuotaFlushIntervalMs) * time.Millisecond
	if iv <= 0 {
		iv = 2 * time.Second
	}
	return &UserPlatformQuotaUsageFlusher{
		cache:        cache,
		quotaRepo:    quotaRepo,
		timingWheel:  tw,
		enabled:      cfg.Database.UserPlatformQuotaFlusherEnabled,
		interval:     iv,
		batchSize:    batch,
		flushTimeout: 3 * time.Second,
		metrics:      &FlusherMetrics{},
	}
}

func (f *UserPlatformQuotaUsageFlusher) Start() {
	if f == nil || !f.enabled {
		return
	}
	f.timingWheel.ScheduleRecurring("deferred:platform_quota", f.interval, f.onTick)
}

func (f *UserPlatformQuotaUsageFlusher) Stop() {
	if f == nil {
		return
	}
	f.stopped.Store(true)
	f.timingWheel.Cancel("deferred:platform_quota")
	f.drainAll()
}

func (f *UserPlatformQuotaUsageFlusher) onTick() {
	if f == nil || f.stopped.Load() {
		return
	}
	f.drainAll()
}

func (f *UserPlatformQuotaUsageFlusher) drainAll() {
	if f == nil {
		return
	}
	root := context.Background()
	for round := 0; round < flusherMaxBatchesPerTick; round++ {
		if !f.drainOneBatch(root) {
			return
		}
	}
	logger.LegacyPrintf("quota_flusher",
		"[QuotaFlusher] hit max batches (%d×%d), dirty set still non-empty, deferring to next tick",
		flusherMaxBatchesPerTick, f.batchSize)
}

func (f *UserPlatformQuotaUsageFlusher) drainOneBatch(parent context.Context) bool {
	ctx, cancel := context.WithTimeout(parent, f.flushTimeout)
	defer cancel()

	keys, err := f.cache.PopDirtyUserPlatformQuotaKeys(ctx, f.batchSize)
	if err != nil {
		f.metrics.FlushErrorTotal.Add(1)
		logger.LegacyPrintf("quota_flusher", "[QuotaFlusher] PopDirty error: %v", err)
		return false
	}
	if len(keys) == 0 {
		return false
	}

	entries, err := f.cache.BatchGetUserPlatformQuotaCache(ctx, keys)
	if err != nil {
		f.metrics.FlushErrorTotal.Add(1)
		f.tryReadd(ctx, keys, "BatchGet")
		logger.LegacyPrintf("quota_flusher", "[QuotaFlusher] BatchGet error: %v", err)
		return false
	}

	snaps := make([]UserPlatformQuotaSnapshot, 0, len(keys))
	for i, key := range keys {
		e := entries[i]
		if e == nil || e.DailyWindowStart == nil || e.WeeklyWindowStart == nil || e.MonthlyWindowStart == nil {
			continue
		}
		snaps = append(snaps, UserPlatformQuotaSnapshot{
			UserID:             key.UserID,
			Platform:           key.Platform,
			DailyUsageUSD:      e.DailyUsageUSD,
			WeeklyUsageUSD:     e.WeeklyUsageUSD,
			MonthlyUsageUSD:    e.MonthlyUsageUSD,
			DailyWindowStart:   *e.DailyWindowStart,
			WeeklyWindowStart:  *e.WeeklyWindowStart,
			MonthlyWindowStart: *e.MonthlyWindowStart,
		})
	}

	if len(snaps) == 0 {
		return len(keys) >= f.batchSize
	}

	t0 := time.Now()
	writeErr := f.quotaRepo.BatchSnapshotUsage(ctx, snaps, time.Now().UTC())
	f.trackMaxLatency(time.Since(t0).Milliseconds())

	if writeErr != nil {
		if errors.Is(writeErr, ErrUserPlatformQuotaFKViolation) {
			f.metrics.FlushFKViolationTotal.Add(1)
			f.metrics.FlushErrorTotal.Add(1)
			logger.LegacyPrintf("quota_flusher", "[QuotaFlusher] FK violation, dropped %d snaps: %v", len(snaps), writeErr)
		} else {
			f.metrics.FlushErrorTotal.Add(1)
			f.tryReadd(ctx, keys, "BatchSnapshot")
			logger.LegacyPrintf("quota_flusher", "[QuotaFlusher] BatchSnapshot error: %v", writeErr)
		}
		return false
	}

	f.metrics.FlushSuccessTotal.Add(1)
	f.metrics.FlushBatchSizeTotal.Add(int64(len(snaps)))
	return len(keys) >= f.batchSize
}

func (f *UserPlatformQuotaUsageFlusher) tryReadd(ctx context.Context, keys []UserPlatformQuotaKey, stage string) {
	if err := f.cache.ReaddDirtyUserPlatformQuotaKeys(ctx, keys); err != nil {
		f.metrics.DirtyLostTotal.Add(int64(len(keys)))
		logger.LegacyPrintf("quota_flusher",
			"[QuotaFlusher] ALERT: readd after %s failed, %d keys lost from dirty set: %v", stage, len(keys), err)
		return
	}
	f.metrics.DirtyReaddTotal.Add(int64(len(keys)))
}

func (f *UserPlatformQuotaUsageFlusher) trackMaxLatency(ms int64) {
	for {
		old := f.metrics.FlushLatencyMsMax.Load()
		if ms <= old || f.metrics.FlushLatencyMsMax.CompareAndSwap(old, ms) {
			return
		}
	}
}
