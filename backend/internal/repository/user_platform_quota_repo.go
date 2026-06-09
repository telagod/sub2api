package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	dbent "github.com/telagod/subme/ent"
	"github.com/telagod/subme/ent/userplatformquota"
	"github.com/telagod/subme/internal/pkg/timezone"
	"github.com/lib/pq"
)

// UserPlatformQuotaRecord is the repository-layer transfer struct,
// decoupled from the ent.UserPlatformQuota entity for use by the service layer.
type UserPlatformQuotaRecord struct {
	UserID             int64
	Platform           string
	DailyLimitUSD      *float64
	WeeklyLimitUSD     *float64
	MonthlyLimitUSD    *float64
	DailyUsageUSD      float64
	WeeklyUsageUSD     float64
	MonthlyUsageUSD    float64
	DailyWindowStart   *time.Time
	WeeklyWindowStart  *time.Time
	MonthlyWindowStart *time.Time
}

// ErrUserPlatformQuotaNotFound is returned by methods that require an existing record.
var ErrUserPlatformQuotaNotFound = fmt.Errorf("user platform quota record not found")

// ErrUserPlatformQuotaFKViolation is returned when a batch upsert contains a
// user_id that does not exist in the users table.
var ErrUserPlatformQuotaFKViolation = errors.New("user platform quota snapshot FK violation")

// UserPlatformQuotaSnapshot represents a Redis window snapshot for absolute-value
// write-back to the database (used by BatchSnapshotUsage).
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

// UserPlatformQuotaRepository defines the data access interface for user platform quotas.
type UserPlatformQuotaRepository interface {
	// BulkInsertInitial idempotently inserts initial quota records (ON CONFLICT DO NOTHING).
	BulkInsertInitial(ctx context.Context, records []UserPlatformQuotaRecord) error
	// GetByUserPlatform queries a single quota record; returns (nil, nil) when not found.
	GetByUserPlatform(ctx context.Context, userID int64, platform string) (*UserPlatformQuotaRecord, error)
	// ListByUser queries all platform quota records for a user (excluding soft-deleted).
	ListByUser(ctx context.Context, userID int64) ([]UserPlatformQuotaRecord, error)
	// IncrementUsageWithReset atomically adds cost, resetting any expired windows first.
	IncrementUsageWithReset(ctx context.Context, userID int64, platform string, cost float64, now time.Time) error
	// ResetExpiredWindow unconditionally resets the specified window's usage and start time.
	ResetExpiredWindow(ctx context.Context, userID int64, platform string, window string, newStart time.Time) error
	// UpsertForUser fully replaces all platform limit configurations for a user.
	UpsertForUser(ctx context.Context, userID int64, records []UserPlatformQuotaRecord) error
	// BatchSnapshotUsage writes an entire batch of usage snapshots as absolute values
	// (not incremental). Shares a single timestamp for created/updated_at.
	// Requires no duplicate (user, platform) pairs within snapshots.
	// FK violations return ErrUserPlatformQuotaFKViolation.
	BatchSnapshotUsage(ctx context.Context, snapshots []UserPlatformQuotaSnapshot, now time.Time) error
}

type userPlatformQuotaRepository struct {
	client *dbent.Client
}

// NewUserPlatformQuotaRepository constructs a UserPlatformQuotaRepository implementation.
func NewUserPlatformQuotaRepository(client *dbent.Client) UserPlatformQuotaRepository {
	return &userPlatformQuotaRepository{client: client}
}

// BulkInsertInitial performs an idempotent batch insert with conditional limit override.
// Only limit_usd and metadata are inserted; usage_usd defaults to 0 and window_start to NULL.
//
// Conflict strategy: existing NULL limits are overwritten with EXCLUDED values while
// non-NULL limits (set by admin via UpsertForUser) are preserved. Usage and window_start
// fields are never touched.
func (repo *userPlatformQuotaRepository) BulkInsertInitial(ctx context.Context, records []UserPlatformQuotaRecord) error {
	if len(records) == 0 {
		return nil
	}

	dbClient := clientFromContext(ctx, repo.client)

	var buf strings.Builder
	_, _ = buf.WriteString("INSERT INTO user_platform_quotas (user_id, platform, daily_limit_usd, weekly_limit_usd, monthly_limit_usd, daily_usage_usd, weekly_usage_usd, monthly_usage_usd, created_at, updated_at) VALUES ")
	params := make([]any, 0, len(records)*6)
	// Use a single timestamp for the entire batch to avoid sub-millisecond drift.
	ts := time.Now()
	for idx, rec := range records {
		offset := idx * 6
		if idx > 0 {
			_, _ = buf.WriteString(",")
		}
		fmt.Fprintf(&buf, "($%d,$%d,$%d,$%d,$%d,0,0,0,$%d,$%d)",
			offset+1, offset+2, offset+3, offset+4, offset+5, offset+6, offset+6)
		params = append(params,
			rec.UserID, rec.Platform,
			rec.DailyLimitUSD, rec.WeeklyLimitUSD, rec.MonthlyLimitUSD,
			ts,
		)
	}
	// Target the partial unique index (deleted_at IS NULL) to avoid conflicts
	// with soft-deleted records. Use COALESCE to preserve existing non-NULL limits.
	_, _ = buf.WriteString(` ON CONFLICT (user_id, platform) WHERE deleted_at IS NULL
		DO UPDATE SET
			daily_limit_usd   = COALESCE(user_platform_quotas.daily_limit_usd, EXCLUDED.daily_limit_usd),
			weekly_limit_usd  = COALESCE(user_platform_quotas.weekly_limit_usd, EXCLUDED.weekly_limit_usd),
			monthly_limit_usd = COALESCE(user_platform_quotas.monthly_limit_usd, EXCLUDED.monthly_limit_usd),
			updated_at        = EXCLUDED.updated_at`)

	_, execErr := dbClient.ExecContext(ctx, buf.String(), params...)
	return execErr
}

// GetByUserPlatform queries a single active quota record via ent.
// Returns (nil, nil) when no matching record exists.
func (repo *userPlatformQuotaRepository) GetByUserPlatform(ctx context.Context, userID int64, platform string) (*UserPlatformQuotaRecord, error) {
	dbClient := clientFromContext(ctx, repo.client)
	row, queryErr := dbClient.UserPlatformQuota.Query().
		Where(
			userplatformquota.UserIDEQ(userID),
			userplatformquota.PlatformEQ(platform),
			userplatformquota.DeletedAtIsNil(),
		).
		Only(ctx)
	if dbent.IsNotFound(queryErr) {
		return nil, nil
	}
	if queryErr != nil {
		return nil, queryErr
	}
	return mapEntQuotaToRecord(row), nil
}

// ListByUser returns all active platform quota records for the given user.
func (repo *userPlatformQuotaRepository) ListByUser(ctx context.Context, userID int64) ([]UserPlatformQuotaRecord, error) {
	dbClient := clientFromContext(ctx, repo.client)
	entities, queryErr := dbClient.UserPlatformQuota.Query().
		Where(
			userplatformquota.UserIDEQ(userID),
			userplatformquota.DeletedAtIsNil(),
		).
		All(ctx)
	if queryErr != nil {
		return nil, queryErr
	}
	result := make([]UserPlatformQuotaRecord, 0, len(entities))
	for _, ent := range entities {
		result = append(result, *mapEntQuotaToRecord(ent))
	}
	return result, nil
}

// IncrementUsageWithReset atomically adds cost to all three windows, resetting
// any expired window before accumulation.
//
// When no record exists (fail-open path), a new row is created with NULL limits
// (unrestricted). The normal registration path (BulkInsertInitial) ensures limits
// are populated at account creation time.
func (repo *userPlatformQuotaRepository) IncrementUsageWithReset(ctx context.Context, userID int64, platform string, cost float64, now time.Time) error {
	return repo.executeInTx(ctx, func(txCtx context.Context, txClient *dbent.Client) error {
		current, queryErr := txClient.UserPlatformQuota.Query().
			Where(
				userplatformquota.UserIDEQ(userID),
				userplatformquota.PlatformEQ(platform),
				userplatformquota.DeletedAtIsNil(),
			).
			ForUpdate().
			Only(txCtx)
		if dbent.IsNotFound(queryErr) {
			// Fail-open: create a new row with NULL limits.
			// Uses ON CONFLICT DO UPDATE to handle concurrent inserts without losing cost.
			const upsertSQL = `INSERT INTO user_platform_quotas
				(user_id, platform, daily_usage_usd, weekly_usage_usd, monthly_usage_usd,
				 daily_window_start, weekly_window_start, monthly_window_start, created_at, updated_at)
				VALUES ($1, $2, $3, $3, $3, $4, $5, $6, $7, $7)
				ON CONFLICT (user_id, platform) WHERE deleted_at IS NULL DO UPDATE SET
					daily_usage_usd   = user_platform_quotas.daily_usage_usd   + EXCLUDED.daily_usage_usd,
					weekly_usage_usd  = user_platform_quotas.weekly_usage_usd  + EXCLUDED.weekly_usage_usd,
					monthly_usage_usd = user_platform_quotas.monthly_usage_usd + EXCLUDED.monthly_usage_usd,
					updated_at        = EXCLUDED.updated_at`
			_, insertErr := txClient.ExecContext(txCtx, upsertSQL,
				userID, platform, cost,
				timezone.StartOfDay(now), timezone.StartOfWeek(now), now, now)
			return insertErr
		}
		if queryErr != nil {
			return queryErr
		}

		updatedDaily := resetOrAccumulate(current.DailyUsageUsd, current.DailyWindowStart, timezone.StartOfDay(now), cost)
		updatedWeekly := resetOrAccumulate(current.WeeklyUsageUsd, current.WeeklyWindowStart, timezone.StartOfWeek(now), cost)
		// Rolling 30-day monthly window: reset when expired, accumulate otherwise.
		updatedMonthly, monthlyStart := rollingMonthlyResetOrAccumulate(current.MonthlyUsageUsd, current.MonthlyWindowStart, cost, now)

		_, saveErr := current.Update().
			SetDailyUsageUsd(updatedDaily).
			SetWeeklyUsageUsd(updatedWeekly).
			SetMonthlyUsageUsd(updatedMonthly).
			SetDailyWindowStart(timezone.StartOfDay(now)).
			SetWeeklyWindowStart(timezone.StartOfWeek(now)).
			SetMonthlyWindowStart(monthlyStart).
			Save(txCtx)
		return saveErr
	})
}

// ResetExpiredWindow unconditionally resets the specified window's usage to zero
// and updates its start time.
//
// WARNING: Despite the name suggesting expiration checking, this method does NOT
// verify whether the window is actually expired. It always resets unconditionally.
// The only intended caller is the admin force-reset endpoint.
//
// Returns ErrUserPlatformQuotaNotFound when no active record matches.
func (repo *userPlatformQuotaRepository) ResetExpiredWindow(ctx context.Context, userID int64, platform string, window string, newStart time.Time) error {
	dbClient := clientFromContext(ctx, repo.client)
	updater := dbClient.UserPlatformQuota.Update().
		Where(
			userplatformquota.UserIDEQ(userID),
			userplatformquota.PlatformEQ(platform),
			userplatformquota.DeletedAtIsNil(),
		)
	switch window {
	case "daily":
		updater = updater.SetDailyUsageUsd(0).SetDailyWindowStart(newStart)
	case "weekly":
		updater = updater.SetWeeklyUsageUsd(0).SetWeeklyWindowStart(newStart)
	case "monthly":
		updater = updater.SetMonthlyUsageUsd(0).SetMonthlyWindowStart(newStart)
	default:
		return fmt.Errorf("unrecognized window type %q", window)
	}
	affected, saveErr := updater.Save(ctx)
	if saveErr != nil {
		return saveErr
	}
	if affected == 0 {
		return ErrUserPlatformQuotaNotFound
	}
	return nil
}

// executeInTx runs fn within a transaction, reusing an existing one from context if available.
func (repo *userPlatformQuotaRepository) executeInTx(ctx context.Context, fn func(txCtx context.Context, txClient *dbent.Client) error) error {
	if existingTx := dbent.TxFromContext(ctx); existingTx != nil {
		return fn(ctx, existingTx.Client())
	}

	txHandle, beginErr := repo.client.Tx(ctx)
	if beginErr != nil {
		return fmt.Errorf("failed to begin quota transaction: %w", beginErr)
	}
	defer func() { _ = txHandle.Rollback() }()

	txCtx := dbent.NewTxContext(ctx, txHandle)
	if execErr := fn(txCtx, txHandle.Client()); execErr != nil {
		return execErr
	}

	if commitErr := txHandle.Commit(); commitErr != nil {
		return fmt.Errorf("failed to commit quota transaction: %w", commitErr)
	}
	return nil
}

// withTx is kept as a method alias.
func (repo *userPlatformQuotaRepository) withTx(ctx context.Context, fn func(txCtx context.Context, txClient *dbent.Client) error) error {
	return repo.executeInTx(ctx, fn)
}

// mapEntQuotaToRecord converts an ent entity to the repository record struct.
func mapEntQuotaToRecord(e *dbent.UserPlatformQuota) *UserPlatformQuotaRecord {
	return &UserPlatformQuotaRecord{
		UserID:             e.UserID,
		Platform:           e.Platform,
		DailyLimitUSD:      e.DailyLimitUsd,
		WeeklyLimitUSD:     e.WeeklyLimitUsd,
		MonthlyLimitUSD:    e.MonthlyLimitUsd,
		DailyUsageUSD:      e.DailyUsageUsd,
		WeeklyUsageUSD:     e.WeeklyUsageUsd,
		MonthlyUsageUSD:    e.MonthlyUsageUsd,
		DailyWindowStart:   e.DailyWindowStart,
		WeeklyWindowStart:  e.WeeklyWindowStart,
		MonthlyWindowStart: e.MonthlyWindowStart,
	}
}

// entQuotaToRecord is kept as a package-level alias.
func entQuotaToRecord(e *dbent.UserPlatformQuota) *UserPlatformQuotaRecord {
	return mapEntQuotaToRecord(e)
}

// resetOrAccumulate returns cost (reset) when the window has changed,
// or prevUsage + cost (accumulate) when the window is still current.
func resetOrAccumulate(prevUsage float64, prevStart *time.Time, currStart time.Time, cost float64) float64 {
	if prevStart == nil || !prevStart.Equal(currStart) {
		return cost
	}
	return prevUsage + cost
}

// maybeReset is kept as a package-level alias.
func maybeReset(prevUsage float64, prevStart *time.Time, currStart time.Time, cost float64) float64 {
	return resetOrAccumulate(prevUsage, prevStart, currStart, cost)
}

// rollingMonthlyResetOrAccumulate handles the 30-day rolling monthly window.
// Expired when prevStart is nil or now - prevStart >= 30 days.
// Returns (newUsage, newWindowStart).
func rollingMonthlyResetOrAccumulate(prevUsage float64, prevStart *time.Time, cost float64, now time.Time) (float64, time.Time) {
	if prevStart == nil || now.Sub(*prevStart) >= 30*24*time.Hour {
		return cost, now
	}
	return prevUsage + cost, *prevStart
}

// monthlyMaybeReset is kept as a package-level alias.
func monthlyMaybeReset(prevUsage float64, prevStart *time.Time, cost float64, now time.Time) (float64, time.Time) {
	return rollingMonthlyResetOrAccumulate(prevUsage, prevStart, cost, now)
}

// UpsertForUser fully replaces all platform limit configurations for a user:
//  1. Soft-deletes active rows for platforms not in the input list
//  2. For each input record, attempts UPDATE first; falls back to INSERT when
//     no active row exists
//
// Only modifies *_limit_usd, deleted_at, and updated_at. Usage and window_start
// fields are preserved.
func (repo *userPlatformQuotaRepository) UpsertForUser(ctx context.Context, userID int64, records []UserPlatformQuotaRecord) error {
	return repo.executeInTx(ctx, func(txCtx context.Context, txClient *dbent.Client) error {
		activePlatforms := make([]string, 0, len(records))
		for _, rec := range records {
			activePlatforms = append(activePlatforms, rec.Platform)
		}
		ts := time.Now()
		if sdErr := markMissingPlatformsDeleted(txCtx, txClient, userID, activePlatforms, ts); sdErr != nil {
			return sdErr
		}
		for _, rec := range records {
			rowsChanged, updErr := overwriteLimitRow(txCtx, txClient, userID, rec, ts)
			if updErr != nil {
				return updErr
			}
			if rowsChanged == 0 {
				if insErr := createLimitRow(txCtx, txClient, userID, rec, ts); insErr != nil {
					return insErr
				}
			}
		}
		return nil
	})
}

// markMissingPlatformsDeleted soft-deletes all active rows for platforms not in
// the keepPlatforms list. When keepPlatforms is empty, all active rows are soft-deleted.
func markMissingPlatformsDeleted(ctx context.Context, client *dbent.Client, userID int64, keepPlatforms []string, ts time.Time) error {
	var stmt string
	var params []any
	if len(keepPlatforms) == 0 {
		stmt = `UPDATE user_platform_quotas SET deleted_at = $2, updated_at = $2
		         WHERE user_id = $1 AND deleted_at IS NULL`
		params = []any{userID, ts}
	} else {
		holders := make([]string, len(keepPlatforms))
		params = make([]any, 0, len(keepPlatforms)+2)
		params = append(params, userID, ts)
		for i, p := range keepPlatforms {
			holders[i] = fmt.Sprintf("$%d", i+3)
			params = append(params, p)
		}
		stmt = fmt.Sprintf(`UPDATE user_platform_quotas SET deleted_at = $2, updated_at = $2
		         WHERE user_id = $1 AND deleted_at IS NULL AND platform NOT IN (%s)`,
			strings.Join(holders, ","))
	}
	_, execErr := client.ExecContext(ctx, stmt, params...)
	return execErr
}

// softDeleteMissingPlatforms is kept as a package-level alias.
func softDeleteMissingPlatforms(ctx context.Context, client *dbent.Client, userID int64, keepPlatforms []string, now time.Time) error {
	return markMissingPlatformsDeleted(ctx, client, userID, keepPlatforms, now)
}

// overwriteLimitRow attempts to UPDATE the active row's limits, returning the
// number of affected rows. Only targets active (non-soft-deleted) rows.
func overwriteLimitRow(ctx context.Context, client *dbent.Client, userID int64, rec UserPlatformQuotaRecord, ts time.Time) (int64, error) {
	const stmt = `UPDATE user_platform_quotas
		SET daily_limit_usd = $1, weekly_limit_usd = $2, monthly_limit_usd = $3,
		    deleted_at = NULL, updated_at = $4
		WHERE user_id = $5 AND platform = $6 AND deleted_at IS NULL`
	execResult, execErr := client.ExecContext(ctx, stmt,
		rec.DailyLimitUSD, rec.WeeklyLimitUSD, rec.MonthlyLimitUSD, ts,
		userID, rec.Platform)
	if execErr != nil {
		return 0, execErr
	}
	return execResult.RowsAffected()
}

// updateLimitsRow is kept as a package-level alias.
func updateLimitsRow(ctx context.Context, client *dbent.Client, userID int64, rec UserPlatformQuotaRecord, now time.Time) (int64, error) {
	return overwriteLimitRow(ctx, client, userID, rec, now)
}

// createLimitRow inserts a new quota row with zero usage. Uses ON CONFLICT DO NOTHING
// to guard against concurrent inserts for the same (user, platform) pair.
// When affected=0 (concurrent insert won), falls back to overwriteLimitRow.
func createLimitRow(ctx context.Context, client *dbent.Client, userID int64, rec UserPlatformQuotaRecord, ts time.Time) error {
	const stmt = `INSERT INTO user_platform_quotas
		(user_id, platform, daily_limit_usd, weekly_limit_usd, monthly_limit_usd,
		 daily_usage_usd, weekly_usage_usd, monthly_usage_usd, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 0, 0, 0, $6, $6)
		ON CONFLICT (user_id, platform) WHERE deleted_at IS NULL DO NOTHING`
	execResult, execErr := client.ExecContext(ctx, stmt,
		userID, rec.Platform,
		rec.DailyLimitUSD, rec.WeeklyLimitUSD, rec.MonthlyLimitUSD,
		ts)
	if execErr != nil {
		return execErr
	}
	rowsInserted, affErr := execResult.RowsAffected()
	if affErr != nil {
		return affErr
	}
	if rowsInserted == 0 {
		// Another concurrent request already created the row; overwrite its limits.
		_, fallbackErr := overwriteLimitRow(ctx, client, userID, rec, ts)
		return fallbackErr
	}
	return nil
}

// insertLimitsRow is kept as a package-level alias.
func insertLimitsRow(ctx context.Context, client *dbent.Client, userID int64, rec UserPlatformQuotaRecord, now time.Time) error {
	return createLimitRow(ctx, client, userID, rec, now)
}

// maxRowsPerBatch is the ceiling for rows in a single BatchSnapshotUsage SQL statement.
// 9 parameters per row x 6000 = 54000, safely under Postgres' 65535 limit.
const maxRowsPerBatch = 6000

// batchRows is kept as a package-level alias.
const batchRows = maxRowsPerBatch

// BatchSnapshotUsage writes an entire batch of usage snapshots as absolute values
// using a multi-row UPSERT. Each batch shares $1=now for timestamps.
//
// Note: batches exceeding maxRowsPerBatch are split into separate SQL statements
// that are NOT wrapped in a single transaction. The caller (flusher) should ensure
// its batch size stays within maxRowsPerBatch.
func (repo *userPlatformQuotaRepository) BatchSnapshotUsage(ctx context.Context, snapshots []UserPlatformQuotaSnapshot, now time.Time) error {
	if len(snapshots) == 0 {
		return nil
	}

	dbClient := clientFromContext(ctx, repo.client)

	for batchStart := 0; batchStart < len(snapshots); batchStart += maxRowsPerBatch {
		batchEnd := batchStart + maxRowsPerBatch
		if batchEnd > len(snapshots) {
			batchEnd = len(snapshots)
		}
		chunk := snapshots[batchStart:batchEnd]

		var buf strings.Builder
		_, _ = buf.WriteString(
			"INSERT INTO user_platform_quotas" +
				" (user_id, platform, daily_usage_usd, weekly_usage_usd, monthly_usage_usd," +
				" daily_window_start, weekly_window_start, monthly_window_start, created_at, updated_at)" +
				" VALUES ")

		// $1 = now (shared); per-row parameters start from $2.
		params := []any{now}
		for i, snap := range chunk {
			if i > 0 {
				_, _ = buf.WriteString(",")
			}
			base := len(params)
			fmt.Fprintf(&buf, "($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$1,$1)",
				base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8)
			params = append(params,
				snap.UserID, snap.Platform,
				snap.DailyUsageUSD, snap.WeeklyUsageUSD, snap.MonthlyUsageUSD,
				snap.DailyWindowStart, snap.WeeklyWindowStart, snap.MonthlyWindowStart,
			)
		}

		_, _ = buf.WriteString(
			" ON CONFLICT (user_id, platform) WHERE deleted_at IS NULL DO UPDATE SET" +
				"  daily_usage_usd      = EXCLUDED.daily_usage_usd," +
				"  weekly_usage_usd     = EXCLUDED.weekly_usage_usd," +
				"  monthly_usage_usd    = EXCLUDED.monthly_usage_usd," +
				"  daily_window_start   = EXCLUDED.daily_window_start," +
				"  weekly_window_start  = EXCLUDED.weekly_window_start," +
				"  monthly_window_start = EXCLUDED.monthly_window_start," +
				"  updated_at           = EXCLUDED.updated_at")

		if _, execErr := dbClient.ExecContext(ctx, buf.String(), params...); execErr != nil {
			var pgErr *pq.Error
			if errors.As(execErr, &pgErr) && pgErr.Code == "23503" {
				return ErrUserPlatformQuotaFKViolation
			}
			return execErr
		}
	}
	return nil
}
