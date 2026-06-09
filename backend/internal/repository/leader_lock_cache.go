package repository

import (
	"context"
	"time"

	"github.com/telagod/subme/internal/service"

	"github.com/redis/go-redis/v9"
)

const leaderLockKeyPrefix = "leader:lock:"

// Lua script for compare-and-delete lock release to prevent a previous holder
// (whose lock expired) from accidentally removing the new owner's lock.
var leaderLockReleaseScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("DEL", KEYS[1])
end
return 0
`)

type leaderLockCache struct {
	rdb *redis.Client
}

// NewLeaderLockCache creates a Redis-backed leader election lock used by
// periodic background jobs to ensure single-instance execution.
func NewLeaderLockCache(rdb *redis.Client) service.LeaderLockCache {
	return &leaderLockCache{rdb: rdb}
}

func (lc *leaderLockCache) TryAcquireLeaderLock(ctx context.Context, lockName, ownerID string, ttl time.Duration) (bool, error) {
	return lc.rdb.SetNX(ctx, leaderLockKeyPrefix+lockName, ownerID, ttl).Result()
}

func (lc *leaderLockCache) ReleaseLeaderLock(ctx context.Context, lockName, ownerID string) error {
	return leaderLockReleaseScript.Run(ctx, lc.rdb, []string{leaderLockKeyPrefix + lockName}, ownerID).Err()
}
