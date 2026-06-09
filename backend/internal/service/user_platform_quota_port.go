package service

import (
	"context"
	"errors"
	"time"
)

// ErrUserPlatformQuotaNotFound is a service-layer sentinel indicating the
// quota record does not exist.
var ErrUserPlatformQuotaNotFound = errors.New("user platform quota not found")

// ErrUserPlatformQuotaFKViolation is a service-layer sentinel for batch
// snapshot UPSERT failures caused by user_id foreign key violations.
var ErrUserPlatformQuotaFKViolation = errors.New("user platform quota snapshot FK violation")

// UserPlatformQuotaSnapshot is the transport struct used by the service-layer
// flusher when writing snapshots to the database.
type UserPlatformQuotaSnapshot struct {
	UserID             int64
	Platform           string
	DailyUsageUSD      float64
	WeeklyUsageUSD     float64
	MonthlyUsageUSD    float64
	DailyWindowStart   time.Time
	WeeklyWindowStart  time.Time
	MonthlyWindowStart time.Time
}

// UserPlatformQuotaRecord is the service-layer DTO decoupled from the
// repository layer.
type UserPlatformQuotaRecord struct {
	UserID          int64
	Platform        string
	DailyLimitUSD   *float64
	WeeklyLimitUSD  *float64
	MonthlyLimitUSD *float64
	DailyUsageUSD   float64
	WeeklyUsageUSD  float64
	MonthlyUsageUSD float64
	DailyWindowStart   *time.Time
	WeeklyWindowStart  *time.Time
	MonthlyWindowStart *time.Time
}

// UserPlatformQuotaRepository defines the data access port for per-user
// per-platform quota management.
type UserPlatformQuotaRepository interface {
	GetByUserPlatform(ctx context.Context, userID int64, platform string) (*UserPlatformQuotaRecord, error)
	BulkInsertInitial(ctx context.Context, records []UserPlatformQuotaRecord) error
	IncrementUsageWithReset(ctx context.Context, userID int64, platform string, cost float64, now time.Time) error
	ListByUser(ctx context.Context, userID int64) ([]UserPlatformQuotaRecord, error)
	UpsertForUser(ctx context.Context, userID int64, records []UserPlatformQuotaRecord) error
	ResetExpiredWindow(ctx context.Context, userID int64, platform string, window string, newStart time.Time) error
	BatchSnapshotUsage(ctx context.Context, snapshots []UserPlatformQuotaSnapshot, now time.Time) error
}
