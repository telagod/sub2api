package repository

import (
	"context"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/telagod/subme/internal/service"
)

const contentModerationFlaggedHashSetKey = "content_moderation:flagged_hashes"

type contentModerationHashCache struct {
	rdb *redis.Client
}

func NewContentModerationHashCache(rdb *redis.Client) service.ContentModerationHashCache {
	return &contentModerationHashCache{rdb: rdb}
}

func (repo *contentModerationHashCache) RecordFlaggedInputHash(ctx context.Context, hash string) error {
	hash = strings.TrimSpace(hash)
	if repo == nil || repo.rdb == nil || hash == "" {
		return nil
	}
	return repo.rdb.SAdd(ctx, contentModerationFlaggedHashSetKey, hash).Err()
}

func (repo *contentModerationHashCache) HasFlaggedInputHash(ctx context.Context, hash string) (bool, error) {
	hash = strings.TrimSpace(hash)
	if repo == nil || repo.rdb == nil || hash == "" {
		return false, nil
	}
	return repo.rdb.SIsMember(ctx, contentModerationFlaggedHashSetKey, hash).Result()
}

func (repo *contentModerationHashCache) DeleteFlaggedInputHash(ctx context.Context, hash string) (bool, error) {
	hash = strings.TrimSpace(hash)
	if repo == nil || repo.rdb == nil || hash == "" {
		return false, nil
	}
	removed, removeErr := repo.rdb.SRem(ctx, contentModerationFlaggedHashSetKey, hash).Result()
	if removeErr != nil {
		return false, removeErr
	}
	return removed > 0, nil
}

func (repo *contentModerationHashCache) ClearFlaggedInputHashes(ctx context.Context) (int64, error) {
	if repo == nil || repo.rdb == nil {
		return 0, nil
	}
	total, countErr := repo.rdb.SCard(ctx, contentModerationFlaggedHashSetKey).Result()
	if countErr != nil {
		return 0, countErr
	}
	if total == 0 {
		return 0, nil
	}
	if delErr := repo.rdb.Del(ctx, contentModerationFlaggedHashSetKey).Err(); delErr != nil {
		return 0, delErr
	}
	return total, nil
}

func (repo *contentModerationHashCache) CountFlaggedInputHashes(ctx context.Context) (int64, error) {
	if repo == nil || repo.rdb == nil {
		return 0, nil
	}
	return repo.rdb.SCard(ctx, contentModerationFlaggedHashSetKey).Result()
}
