package service

import (
	"context"
	"database/sql"
	"time"
)

// LeaderLockCache provides distributed mutual exclusion for periodic background
// jobs. Implemented in the repository layer (typically Redis-backed) so the
// service layer has no direct Redis dependency. Release uses compare-and-delete
// keyed by owner to prevent a stale holder from removing a peer's lock.
type LeaderLockCache interface {
	// TryAcquireLeaderLock sets key=owner with the given TTL if the key does
	// not already exist. Returns true when the caller becomes the owner.
	TryAcquireLeaderLock(ctx context.Context, key, owner string, ttl time.Duration) (bool, error)
	// ReleaseLeaderLock removes the key only if it is still held by owner.
	ReleaseLeaderLock(ctx context.Context, key, owner string) error
}

// tryAcquireSingletonLeaderLock attempts to acquire a distributed leader lock
// for a periodic background job. It tries the Redis-backed cache first, then
// falls back to a Postgres advisory lock when the cache is unavailable.
//
// Returns:
//   - (releaseFunc, true) when the lock is acquired; caller must defer releaseFunc
//   - (nil, false) when another instance holds the lock
//   - (no-op func, true) when no coordination backend is configured, so the
//     job always runs (suitable for single-instance deployments or tests)
func tryAcquireSingletonLeaderLock(ctx context.Context, cache LeaderLockCache, db *sql.DB, key, owner string, ttl time.Duration) (func(), bool) {
	if ctx == nil {
		ctx = context.Background()
	}

	// Prefer the distributed cache when available.
	if cache != nil {
		acquired, err := cache.TryAcquireLeaderLock(ctx, key, owner, ttl)
		if err == nil {
			if !acquired {
				return nil, false
			}
			cleanup := func() {
				bg, done := context.WithTimeout(context.Background(), 2*time.Second)
				defer done()
				_ = cache.ReleaseLeaderLock(bg, key, owner)
			}
			return cleanup, true
		}
		// On cache failure, fall through to DB advisory lock to avoid
		// stampeding the job across all instances.
	}

	// Try database advisory lock as secondary coordination mechanism.
	if db != nil {
		return tryAcquireDBAdvisoryLock(ctx, db, hashAdvisoryLockID(key))
	}

	// No backend configured: allow the job to proceed ungated.
	return func() {}, true
}
