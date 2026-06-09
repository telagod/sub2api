package repository

import (
	"context"
	"fmt"

	"github.com/telagod/subme/internal/service"
	"github.com/redis/go-redis/v9"
)

const openAI403CounterPrefix = "openai_403_count:account:"

var openAI403CounterIncrScript = redis.NewScript(`
	local k = KEYS[1]
	local expiry = tonumber(ARGV[1])

	local n = redis.call('INCR', k)
	if n == 1 then
		redis.call('EXPIRE', k, expiry)
	end

	return n
`)

type openAI403CounterCache struct {
	rdb *redis.Client
}

func NewOpenAI403CounterCache(rdb *redis.Client) service.OpenAI403CounterCache {
	return &openAI403CounterCache{rdb: rdb}
}

func (cc *openAI403CounterCache) IncrementOpenAI403Count(ctx context.Context, accountID int64, windowMinutes int) (int64, error) {
	redisKey := fmt.Sprintf("%s%d", openAI403CounterPrefix, accountID)

	ttlSec := windowMinutes * 60
	if ttlSec < 60 {
		ttlSec = 60
	}

	count, execErr := openAI403CounterIncrScript.Run(ctx, cc.rdb, []string{redisKey}, ttlSec).Int64()
	if execErr != nil {
		return 0, fmt.Errorf("failed to increment 403 counter: %w", execErr)
	}
	return count, nil
}

func (cc *openAI403CounterCache) ResetOpenAI403Count(ctx context.Context, accountID int64) error {
	redisKey := fmt.Sprintf("%s%d", openAI403CounterPrefix, accountID)
	return cc.rdb.Del(ctx, redisKey).Err()
}
