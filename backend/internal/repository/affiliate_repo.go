package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	dbent "github.com/telagod/subme/ent"
	"github.com/telagod/subme/ent/user"
	"github.com/telagod/subme/internal/service"
	"github.com/lib/pq"
)

const (
	affiliateCodeLength      = 12
	affiliateCodeMaxAttempts = 12
)

var affiliateCodeCharset = []byte("ABCDEFGHJKLMNPQRSTUVWXYZ23456789")

const affiliateUserOverviewSQL = `
SELECT ua.user_id,
       COALESCE(u.email, ''),
       COALESCE(u.username, ''),
       ua.aff_code,
       COALESCE(ua.aff_rebate_rate_percent, 0)::double precision,
       (ua.aff_rebate_rate_percent IS NOT NULL) AS has_custom_rate,
       ua.aff_count,
       COALESCE(rebated.rebated_invitee_count, 0),
       (ua.aff_quota + COALESCE(matured.matured_frozen_quota, 0))::double precision,
       ua.aff_history_quota::double precision
FROM user_affiliates ua
JOIN users u ON u.id = ua.user_id
LEFT JOIN (
    SELECT user_id, COUNT(DISTINCT source_user_id)::integer AS rebated_invitee_count
    FROM user_affiliate_ledger
    WHERE action = 'accrue' AND source_user_id IS NOT NULL
    GROUP BY user_id
) rebated ON rebated.user_id = ua.user_id
LEFT JOIN (
    SELECT user_id, COALESCE(SUM(amount), 0)::double precision AS matured_frozen_quota
    FROM user_affiliate_ledger
    WHERE action = 'accrue' AND frozen_until IS NOT NULL AND frozen_until <= NOW()
    GROUP BY user_id
) matured ON matured.user_id = ua.user_id
WHERE ua.user_id = $1
LIMIT 1`

type affiliateQueryRunner interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type affiliateRepository struct {
	client *dbent.Client
}

func NewAffiliateRepository(client *dbent.Client, _ *sql.DB) service.AffiliateRepository {
	return &affiliateRepository{client: client}
}

func (r *affiliateRepository) EnsureUserAffiliate(ctx context.Context, userID int64) (*service.AffiliateSummary, error) {
	if userID <= 0 {
		return nil, service.ErrUserNotFound
	}
	db := clientFromContext(ctx, r.client)
	return provisionUserAffiliate(ctx, db, userID)
}

func (r *affiliateRepository) GetAffiliateByCode(ctx context.Context, code string) (*service.AffiliateSummary, error) {
	db := clientFromContext(ctx, r.client)
	return fetchAffiliateByCode(ctx, db, code)
}

func (r *affiliateRepository) BindInviter(ctx context.Context, userID, inviterID int64) (bool, error) {
	linked := false
	txErr := r.runInTx(ctx, func(txCtx context.Context, txDB *dbent.Client) error {
		if _, err := provisionUserAffiliate(txCtx, txDB, userID); err != nil {
			return err
		}
		if _, err := provisionUserAffiliate(txCtx, txDB, inviterID); err != nil {
			return err
		}

		result, execErr := txDB.ExecContext(txCtx,
			"UPDATE user_affiliates SET inviter_id = $1, updated_at = NOW() WHERE user_id = $2 AND inviter_id IS NULL",
			inviterID, userID,
		)
		if execErr != nil {
			return fmt.Errorf("set inviter for user: %w", execErr)
		}
		n, _ := result.RowsAffected()
		if n == 0 {
			linked = false
			return nil
		}

		if _, execErr = txDB.ExecContext(txCtx,
			"UPDATE user_affiliates SET aff_count = aff_count + 1, updated_at = NOW() WHERE user_id = $1",
			inviterID,
		); execErr != nil {
			return fmt.Errorf("bump inviter aff_count: %w", execErr)
		}
		linked = true
		return nil
	})
	if txErr != nil {
		return false, txErr
	}
	return linked, nil
}

func (r *affiliateRepository) AccrueQuota(ctx context.Context, inviterID, inviteeUserID int64, amount float64, freezeHours int, sourceOrderID *int64) (bool, error) {
	if amount <= 0 {
		return false, nil
	}

	credited := false
	txErr := r.runInTx(ctx, func(txCtx context.Context, txDB *dbent.Client) error {
		// freezeHours > 0: add to frozen quota; == 0: add to available quota directly
		var stmt string
		if freezeHours > 0 {
			stmt = "UPDATE user_affiliates SET aff_frozen_quota = aff_frozen_quota + $1, aff_history_quota = aff_history_quota + $1, updated_at = NOW() WHERE user_id = $2"
		} else {
			stmt = "UPDATE user_affiliates SET aff_quota = aff_quota + $1, aff_history_quota = aff_history_quota + $1, updated_at = NOW() WHERE user_id = $2"
		}
		result, execErr := txDB.ExecContext(txCtx, stmt, amount, inviterID)
		if execErr != nil {
			return execErr
		}
		n, _ := result.RowsAffected()
		if n == 0 {
			credited = false
			return nil
		}

		if freezeHours > 0 {
			if _, execErr = txDB.ExecContext(txCtx, `
INSERT INTO user_affiliate_ledger (user_id, action, amount, source_user_id, source_order_id, frozen_until, created_at, updated_at)
VALUES ($1, 'accrue', $2, $3, $4, NOW() + make_interval(hours => $5), NOW(), NOW())`,
				inviterID, amount, inviteeUserID, optionalInt64Param(sourceOrderID), freezeHours); execErr != nil {
				return fmt.Errorf("record frozen accrue ledger entry: %w", execErr)
			}
		} else {
			if _, execErr = txDB.ExecContext(txCtx, `
INSERT INTO user_affiliate_ledger (user_id, action, amount, source_user_id, source_order_id, created_at, updated_at)
VALUES ($1, 'accrue', $2, $3, $4, NOW(), NOW())`, inviterID, amount, inviteeUserID, optionalInt64Param(sourceOrderID)); execErr != nil {
				return fmt.Errorf("record accrue ledger entry: %w", execErr)
			}
		}

		credited = true
		return nil
	})
	if txErr != nil {
		return false, txErr
	}
	return credited, nil
}

func (r *affiliateRepository) GetAccruedRebateFromInvitee(ctx context.Context, inviterID, inviteeUserID int64) (float64, error) {
	db := clientFromContext(ctx, r.client)
	rows, queryErr := db.QueryContext(ctx,
		`SELECT COALESCE(SUM(amount), 0)::double precision FROM user_affiliate_ledger WHERE user_id = $1 AND source_user_id = $2 AND action = 'accrue'`,
		inviterID, inviteeUserID)
	if queryErr != nil {
		return 0, fmt.Errorf("sum accrued rebate for invitee: %w", queryErr)
	}
	defer func() { _ = rows.Close() }()
	var sum float64
	if rows.Next() {
		if scanErr := rows.Scan(&sum); scanErr != nil {
			return 0, scanErr
		}
	}
	return sum, rows.Close()
}

func (r *affiliateRepository) ThawFrozenQuota(ctx context.Context, userID int64) (float64, error) {
	var released float64
	txErr := r.runInTx(ctx, func(txCtx context.Context, txDB *dbent.Client) error {
		var err error
		released, err = thawMaturedQuota(txCtx, txDB, userID)
		return err
	})
	return released, txErr
}

// thawMaturedQuota moves matured frozen quota to available quota within an existing tx.
func thawMaturedQuota(txCtx context.Context, txDB *dbent.Client, userID int64) (float64, error) {
	rows, queryErr := txDB.QueryContext(txCtx, `
WITH matured AS (
    UPDATE user_affiliate_ledger
    SET frozen_until = NULL, updated_at = NOW()
    WHERE user_id = $1
      AND frozen_until IS NOT NULL
      AND frozen_until <= NOW()
    RETURNING amount
)
SELECT COALESCE(SUM(amount), 0) FROM matured`, userID)
	if queryErr != nil {
		return 0, fmt.Errorf("collect matured frozen entries: %w", queryErr)
	}
	defer func() { _ = rows.Close() }()

	var total float64
	if rows.Next() {
		if scanErr := rows.Scan(&total); scanErr != nil {
			return 0, scanErr
		}
	}
	if closeErr := rows.Close(); closeErr != nil {
		return 0, closeErr
	}
	if total <= 0 {
		return 0, nil
	}

	_, execErr := txDB.ExecContext(txCtx, `
UPDATE user_affiliates
SET aff_quota = aff_quota + $1,
    aff_frozen_quota = GREATEST(aff_frozen_quota - $1, 0),
    updated_at = NOW()
WHERE user_id = $2`, total, userID)
	if execErr != nil {
		return 0, fmt.Errorf("shift thawed quota to available: %w", execErr)
	}
	return total, nil
}

func (r *affiliateRepository) TransferQuotaToBalance(ctx context.Context, userID int64) (float64, float64, error) {
	var moved float64
	var resultBalance float64

	txErr := r.runInTx(ctx, func(txCtx context.Context, txDB *dbent.Client) error {
		if _, err := provisionUserAffiliate(txCtx, txDB, userID); err != nil {
			return err
		}

		// Thaw any matured frozen quota first.
		if _, err := thawMaturedQuota(txCtx, txDB, userID); err != nil {
			return fmt.Errorf("pre-transfer thaw: %w", err)
		}

		rows, queryErr := txDB.QueryContext(txCtx, `
WITH claimed AS (
	SELECT aff_quota::double precision AS amount
	FROM user_affiliates
	WHERE user_id = $1
	  AND aff_quota > 0
	FOR UPDATE
),
cleared AS (
	UPDATE user_affiliates ua
	SET aff_quota = 0,
	    updated_at = NOW()
	FROM claimed c
	WHERE ua.user_id = $1
	RETURNING c.amount
)
SELECT amount
FROM cleared`, userID)
		if queryErr != nil {
			return fmt.Errorf("claim available affiliate quota: %w", queryErr)
		}

		if !rows.Next() {
			_ = rows.Close()
			if err := rows.Err(); err != nil {
				return err
			}
			return service.ErrAffiliateQuotaEmpty
		}
		if scanErr := rows.Scan(&moved); scanErr != nil {
			_ = rows.Close()
			return scanErr
		}
		if closeErr := rows.Close(); closeErr != nil {
			return closeErr
		}
		if moved <= 0 {
			return service.ErrAffiliateQuotaEmpty
		}

		n, updateErr := txDB.User.Update().
			Where(user.IDEQ(userID)).
			AddBalance(moved).
			AddTotalRecharged(moved).
			Save(txCtx)
		if updateErr != nil {
			return fmt.Errorf("credit user balance with affiliate quota: %w", updateErr)
		}
		if n == 0 {
			return service.ErrUserNotFound
		}

		var balErr error
		resultBalance, balErr = readUserBalance(txCtx, txDB, userID)
		if balErr != nil {
			return balErr
		}

		snap, snapErr := captureTransferSnapshot(txCtx, txDB, userID)
		if snapErr != nil {
			return snapErr
		}

		if _, execErr := txDB.ExecContext(txCtx, `
INSERT INTO user_affiliate_ledger (
    user_id,
    action,
    amount,
    source_user_id,
    balance_after,
    aff_quota_after,
    aff_frozen_quota_after,
    aff_history_quota_after,
    created_at,
    updated_at
)
VALUES ($1, 'transfer', $2, NULL, $3, $4, $5, $6, NOW(), NOW())`,
			userID,
			moved,
			snap.BalanceAfter,
			snap.AvailableQuotaAfter,
			snap.FrozenQuotaAfter,
			snap.HistoryQuotaAfter,
		); execErr != nil {
			return fmt.Errorf("record transfer ledger entry: %w", execErr)
		}

		return nil
	})
	if txErr != nil {
		return 0, 0, txErr
	}

	return moved, resultBalance, nil
}

func (r *affiliateRepository) ListInvitees(ctx context.Context, inviterID int64, limit int) ([]service.AffiliateInvitee, error) {
	if limit <= 0 {
		limit = 100
	}
	db := clientFromContext(ctx, r.client)
	rows, queryErr := db.QueryContext(ctx, `
SELECT ua.user_id,
       COALESCE(u.email, ''),
       COALESCE(u.username, ''),
       ua.created_at,
       COALESCE(SUM(ual.amount), 0)::double precision AS total_rebate
FROM user_affiliates ua
LEFT JOIN users u ON u.id = ua.user_id
LEFT JOIN user_affiliate_ledger ual
       ON ual.user_id = $1
      AND ual.source_user_id = ua.user_id
      AND ual.action = 'accrue'
WHERE ua.inviter_id = $1
GROUP BY ua.user_id, u.email, u.username, ua.created_at
ORDER BY ua.created_at DESC
LIMIT $2`, inviterID, limit)
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { _ = rows.Close() }()

	list := make([]service.AffiliateInvitee, 0)
	for rows.Next() {
		var entry service.AffiliateInvitee
		var ts time.Time
		if scanErr := rows.Scan(&entry.UserID, &entry.Email, &entry.Username, &ts, &entry.TotalRebate); scanErr != nil {
			return nil, scanErr
		}
		entry.CreatedAt = &ts
		list = append(list, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *affiliateRepository) ListAffiliateInviteRecords(ctx context.Context, filter service.AffiliateRecordFilter) ([]service.AffiliateInviteRecord, int64, error) {
	db := clientFromContext(ctx, r.client)
	whereClause, params := composeRecordFilter(filter, "ua.created_at", []string{
		"inviter.email", "inviter.username", "invitee.email", "invitee.username",
		"ua.inviter_id::text", "ua.user_id::text", "inviter_aff.aff_code",
	})

	cnt, cntErr := countRecords(ctx, db, `
SELECT COUNT(*)
FROM user_affiliates ua
JOIN users invitee ON invitee.id = ua.user_id
JOIN users inviter ON inviter.id = ua.inviter_id
JOIN user_affiliates inviter_aff ON inviter_aff.user_id = ua.inviter_id
`+whereClause, params...)
	if cntErr != nil {
		return nil, 0, cntErr
	}

	sorting := composeRecordSort(filter, map[string]string{
		"inviter":      "inviter.email",
		"invitee":      "invitee.email",
		"aff_code":     "inviter_aff.aff_code",
		"total_rebate": "total_rebate",
		"created_at":   "ua.created_at",
	}, "ua.created_at")
	params = append(params, filter.PageSize, (filter.Page-1)*filter.PageSize)
	rows, queryErr := db.QueryContext(ctx, `
SELECT ua.inviter_id,
       COALESCE(inviter.email, ''),
       COALESCE(inviter.username, ''),
       ua.user_id,
       COALESCE(invitee.email, ''),
       COALESCE(invitee.username, ''),
       COALESCE(inviter_aff.aff_code, ''),
       COALESCE(SUM(ual.amount), 0)::double precision AS total_rebate,
       ua.created_at
FROM user_affiliates ua
JOIN users invitee ON invitee.id = ua.user_id
JOIN users inviter ON inviter.id = ua.inviter_id
JOIN user_affiliates inviter_aff ON inviter_aff.user_id = ua.inviter_id
LEFT JOIN user_affiliate_ledger ual
       ON ual.user_id = ua.inviter_id
      AND ual.source_user_id = ua.user_id
      AND ual.action = 'accrue'
`+whereClause+`
GROUP BY ua.inviter_id, inviter.email, inviter.username, ua.user_id, invitee.email, invitee.username, inviter_aff.aff_code, ua.created_at
`+sorting+`
LIMIT $`+fmt.Sprint(len(params)-1)+` OFFSET $`+fmt.Sprint(len(params)), params...)
	if queryErr != nil {
		return nil, 0, queryErr
	}
	defer func() { _ = rows.Close() }()

	records := make([]service.AffiliateInviteRecord, 0)
	for rows.Next() {
		var rec service.AffiliateInviteRecord
		if scanErr := rows.Scan(
			&rec.InviterID,
			&rec.InviterEmail,
			&rec.InviterUsername,
			&rec.InviteeID,
			&rec.InviteeEmail,
			&rec.InviteeUsername,
			&rec.AffCode,
			&rec.TotalRebate,
			&rec.CreatedAt,
		); scanErr != nil {
			return nil, 0, scanErr
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return records, cnt, nil
}

func (r *affiliateRepository) ListAffiliateRebateRecords(ctx context.Context, filter service.AffiliateRecordFilter) ([]service.AffiliateRebateRecord, int64, error) {
	db := clientFromContext(ctx, r.client)
	whereClause, params := composeRecordFilter(filter, "ual.created_at", []string{
		"inviter.email", "inviter.username", "invitee.email", "invitee.username",
		"po.id::text", "po.out_trade_no", "po.payment_type", "po.status",
	})
	fromJoin := `
FROM user_affiliate_ledger ual
JOIN payment_orders po ON po.id = ual.source_order_id
JOIN users invitee ON invitee.id = ual.source_user_id
JOIN users inviter ON inviter.id = ual.user_id
WHERE ual.action = 'accrue'
  AND ual.source_order_id IS NOT NULL`
	if whereClause != "" {
		whereClause = strings.Replace(whereClause, "WHERE ", " AND ", 1)
	}

	cnt, cntErr := countRecords(ctx, db, "SELECT COUNT(*) "+fromJoin+whereClause, params...)
	if cntErr != nil {
		return nil, 0, cntErr
	}

	sorting := composeRecordSort(filter, map[string]string{
		"order":         "po.id",
		"inviter":       "inviter.email",
		"invitee":       "invitee.email",
		"order_amount":  "po.amount",
		"pay_amount":    "po.pay_amount",
		"rebate_amount": "ual.amount",
		"payment_type":  "po.payment_type",
		"order_status":  "po.status",
		"created_at":    "ual.created_at",
	}, "ual.created_at")
	params = append(params, filter.PageSize, (filter.Page-1)*filter.PageSize)
	rows, queryErr := db.QueryContext(ctx, `
SELECT po.id,
       po.out_trade_no,
       ual.user_id,
       COALESCE(inviter.email, ''),
       COALESCE(inviter.username, ''),
       ual.source_user_id,
       COALESCE(invitee.email, ''),
       COALESCE(invitee.username, ''),
       po.amount::double precision,
       po.pay_amount::double precision,
       ual.amount::double precision,
       po.payment_type,
       po.status,
       ual.created_at
`+fromJoin+whereClause+`
`+sorting+`
LIMIT $`+fmt.Sprint(len(params)-1)+` OFFSET $`+fmt.Sprint(len(params)), params...)
	if queryErr != nil {
		return nil, 0, queryErr
	}
	defer func() { _ = rows.Close() }()

	records := make([]service.AffiliateRebateRecord, 0)
	for rows.Next() {
		var rec service.AffiliateRebateRecord
		if scanErr := rows.Scan(
			&rec.OrderID,
			&rec.OutTradeNo,
			&rec.InviterID,
			&rec.InviterEmail,
			&rec.InviterUsername,
			&rec.InviteeID,
			&rec.InviteeEmail,
			&rec.InviteeUsername,
			&rec.OrderAmount,
			&rec.PayAmount,
			&rec.RebateAmount,
			&rec.PaymentType,
			&rec.OrderStatus,
			&rec.CreatedAt,
		); scanErr != nil {
			return nil, 0, scanErr
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return records, cnt, nil
}

func (r *affiliateRepository) ListAffiliateTransferRecords(ctx context.Context, filter service.AffiliateRecordFilter) ([]service.AffiliateTransferRecord, int64, error) {
	db := clientFromContext(ctx, r.client)
	whereClause, params := composeRecordFilter(filter, "ual.created_at", []string{
		"u.email", "u.username", "u.id::text",
	})
	fromJoin := `
FROM user_affiliate_ledger ual
JOIN users u ON u.id = ual.user_id
WHERE ual.action = 'transfer'`
	if whereClause != "" {
		whereClause = strings.Replace(whereClause, "WHERE ", " AND ", 1)
	}

	cnt, cntErr := countRecords(ctx, db, "SELECT COUNT(*) "+fromJoin+whereClause, params...)
	if cntErr != nil {
		return nil, 0, cntErr
	}

	sorting := composeRecordSort(filter, map[string]string{
		"user":                  "u.email",
		"amount":                "ual.amount",
		"balance_after":         "ual.balance_after",
		"available_quota_after": "ual.aff_quota_after",
		"frozen_quota_after":    "ual.aff_frozen_quota_after",
		"history_quota_after":   "ual.aff_history_quota_after",
		"created_at":            "ual.created_at",
	}, "ual.created_at")
	params = append(params, filter.PageSize, (filter.Page-1)*filter.PageSize)
	rows, queryErr := db.QueryContext(ctx, `
SELECT ual.id,
       ual.user_id,
       COALESCE(u.email, ''),
       COALESCE(u.username, ''),
       ual.amount::double precision,
       ual.balance_after::double precision,
       ual.aff_quota_after::double precision,
       ual.aff_frozen_quota_after::double precision,
       ual.aff_history_quota_after::double precision,
       ual.created_at
`+fromJoin+whereClause+`
`+sorting+`
LIMIT $`+fmt.Sprint(len(params)-1)+` OFFSET $`+fmt.Sprint(len(params)), params...)
	if queryErr != nil {
		return nil, 0, queryErr
	}
	defer func() { _ = rows.Close() }()

	records := make([]service.AffiliateTransferRecord, 0)
	for rows.Next() {
		var rec service.AffiliateTransferRecord
		var balAfter sql.NullFloat64
		var availAfter sql.NullFloat64
		var frozenAfter sql.NullFloat64
		var histAfter sql.NullFloat64
		if scanErr := rows.Scan(
			&rec.LedgerID,
			&rec.UserID,
			&rec.UserEmail,
			&rec.Username,
			&rec.Amount,
			&balAfter,
			&availAfter,
			&frozenAfter,
			&histAfter,
			&rec.CreatedAt,
		); scanErr != nil {
			return nil, 0, scanErr
		}
		rec.BalanceAfter = toFloat64Ptr(balAfter)
		rec.AvailableQuotaAfter = toFloat64Ptr(availAfter)
		rec.FrozenQuotaAfter = toFloat64Ptr(frozenAfter)
		rec.HistoryQuotaAfter = toFloat64Ptr(histAfter)
		rec.SnapshotAvailable = balAfter.Valid &&
			availAfter.Valid &&
			frozenAfter.Valid &&
			histAfter.Valid
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return records, cnt, nil
}

func (r *affiliateRepository) GetAffiliateUserOverview(ctx context.Context, userID int64) (*service.AffiliateUserOverview, error) {
	if userID <= 0 {
		return nil, service.ErrUserNotFound
	}
	db := clientFromContext(ctx, r.client)
	rows, queryErr := db.QueryContext(ctx, affiliateUserOverviewSQL, userID)
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { _ = rows.Close() }()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, service.ErrUserNotFound
	}

	var ov service.AffiliateUserOverview
	var rate float64
	var rateSet bool
	if scanErr := rows.Scan(
		&ov.UserID,
		&ov.Email,
		&ov.Username,
		&ov.AffCode,
		&rate,
		&rateSet,
		&ov.InvitedCount,
		&ov.RebatedInviteeCount,
		&ov.AvailableQuota,
		&ov.HistoryQuota,
	); scanErr != nil {
		return nil, scanErr
	}
	if rateSet {
		ov.RebateRatePercent = rate
		ov.RebateRateCustom = true
	}
	return &ov, rows.Err()
}

func composeRecordFilter(filter service.AffiliateRecordFilter, timeCol string, searchCols []string) (string, []any) {
	conditions := make([]string, 0, 3)
	params := make([]any, 0, 3)
	if filter.StartAt != nil {
		params = append(params, *filter.StartAt)
		conditions = append(conditions, fmt.Sprintf("%s >= $%d", timeCol, len(params)))
	}
	if filter.EndAt != nil {
		params = append(params, *filter.EndAt)
		conditions = append(conditions, fmt.Sprintf("%s <= $%d", timeCol, len(params)))
	}
	keyword := strings.TrimSpace(filter.Search)
	if keyword != "" && len(searchCols) > 0 {
		params = append(params, "%"+strings.ToLower(keyword)+"%")
		likes := make([]string, 0, len(searchCols))
		for _, col := range searchCols {
			likes = append(likes, fmt.Sprintf("LOWER(%s) LIKE $%d", col, len(params)))
		}
		conditions = append(conditions, "("+strings.Join(likes, " OR ")+")")
	}
	if len(conditions) == 0 {
		return "", params
	}
	return "WHERE " + strings.Join(conditions, " AND "), params
}

func composeRecordSort(filter service.AffiliateRecordFilter, columnMap map[string]string, defaultCol string) string {
	col := columnMap[filter.SortBy]
	if col == "" {
		col = defaultCol
	}
	dir := "DESC"
	if !filter.SortDesc {
		dir = "ASC"
	}
	return "ORDER BY " + col + " " + dir + " NULLS LAST"
}

func countRecords(ctx context.Context, db affiliateQueryRunner, query string, args ...any) (int64, error) {
	rows, queryErr := db.QueryContext(ctx, query, args...)
	if queryErr != nil {
		return 0, queryErr
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return 0, rows.Err()
	}
	var n int64
	if scanErr := rows.Scan(&n); scanErr != nil {
		return 0, scanErr
	}
	return n, rows.Err()
}

func (r *affiliateRepository) runInTx(ctx context.Context, fn func(txCtx context.Context, txDB *dbent.Client) error) error {
	if existing := dbent.TxFromContext(ctx); existing != nil {
		return fn(ctx, existing.Client())
	}

	tx, txErr := r.client.Tx(ctx)
	if txErr != nil {
		return fmt.Errorf("start affiliate transaction: %w", txErr)
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := dbent.NewTxContext(ctx, tx)
	if execErr := fn(txCtx, tx.Client()); execErr != nil {
		return execErr
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("commit affiliate transaction: %w", commitErr)
	}
	return nil
}

func provisionUserAffiliate(ctx context.Context, db affiliateQueryRunner, userID int64) (*service.AffiliateSummary, error) {
	existing, readErr := fetchAffiliateByUserID(ctx, db, userID)
	if readErr == nil {
		return existing, nil
	}
	if !errors.Is(readErr, service.ErrAffiliateProfileNotFound) {
		return nil, readErr
	}

	for attempt := 0; attempt < affiliateCodeMaxAttempts; attempt++ {
		code, genErr := mintAffiliateCode()
		if genErr != nil {
			return nil, genErr
		}
		_, insertErr := db.ExecContext(ctx, `
INSERT INTO user_affiliates (user_id, aff_code, created_at, updated_at)
VALUES ($1, $2, NOW(), NOW())
ON CONFLICT (user_id) DO NOTHING`, userID, code)
		if insertErr == nil {
			break
		}
		if isUniqueConstraintViolation(insertErr) {
			continue
		}
		return nil, insertErr
	}

	return fetchAffiliateByUserID(ctx, db, userID)
}

func fetchAffiliateByUserID(ctx context.Context, db affiliateQueryRunner, userID int64) (*service.AffiliateSummary, error) {
	rows, queryErr := db.QueryContext(ctx, `
SELECT user_id,
       aff_code,
       aff_code_custom,
       aff_rebate_rate_percent,
       inviter_id,
       aff_count,
       aff_quota::double precision,
       aff_frozen_quota::double precision,
       aff_history_quota::double precision,
       created_at,
       updated_at
FROM user_affiliates
WHERE user_id = $1`, userID)
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, service.ErrAffiliateProfileNotFound
	}

	var s service.AffiliateSummary
	var inviter sql.NullInt64
	var rate sql.NullFloat64
	if scanErr := rows.Scan(
		&s.UserID,
		&s.AffCode,
		&s.AffCodeCustom,
		&rate,
		&inviter,
		&s.AffCount,
		&s.AffQuota,
		&s.AffFrozenQuota,
		&s.AffHistoryQuota,
		&s.CreatedAt,
		&s.UpdatedAt,
	); scanErr != nil {
		return nil, scanErr
	}
	if inviter.Valid {
		s.InviterID = &inviter.Int64
	}
	if rate.Valid {
		v := rate.Float64
		s.AffRebateRatePercent = &v
	}
	return &s, nil
}

func fetchAffiliateByCode(ctx context.Context, db affiliateQueryRunner, code string) (*service.AffiliateSummary, error) {
	rows, queryErr := db.QueryContext(ctx, `
SELECT user_id,
       aff_code,
       aff_code_custom,
       aff_rebate_rate_percent,
       inviter_id,
       aff_count,
       aff_quota::double precision,
       aff_frozen_quota::double precision,
       aff_history_quota::double precision,
       created_at,
       updated_at
FROM user_affiliates
WHERE aff_code = $1
LIMIT 1`, strings.ToUpper(strings.TrimSpace(code)))
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { _ = rows.Close() }()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, service.ErrAffiliateProfileNotFound
	}

	var s service.AffiliateSummary
	var inviter sql.NullInt64
	var rate sql.NullFloat64
	if scanErr := rows.Scan(
		&s.UserID,
		&s.AffCode,
		&s.AffCodeCustom,
		&rate,
		&inviter,
		&s.AffCount,
		&s.AffQuota,
		&s.AffFrozenQuota,
		&s.AffHistoryQuota,
		&s.CreatedAt,
		&s.UpdatedAt,
	); scanErr != nil {
		return nil, scanErr
	}
	if inviter.Valid {
		s.InviterID = &inviter.Int64
	}
	if rate.Valid {
		v := rate.Float64
		s.AffRebateRatePercent = &v
	}
	return &s, nil
}

func readUserBalance(ctx context.Context, db affiliateQueryRunner, userID int64) (float64, error) {
	rows, queryErr := db.QueryContext(ctx,
		"SELECT balance::double precision FROM users WHERE id = $1 LIMIT 1",
		userID,
	)
	if queryErr != nil {
		return 0, queryErr
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return 0, err
		}
		return 0, service.ErrUserNotFound
	}
	var bal float64
	if scanErr := rows.Scan(&bal); scanErr != nil {
		return 0, scanErr
	}
	return bal, nil
}

type affiliateTransferSnapshot struct {
	BalanceAfter        float64
	AvailableQuotaAfter float64
	FrozenQuotaAfter    float64
	HistoryQuotaAfter   float64
}

func captureTransferSnapshot(ctx context.Context, db affiliateQueryRunner, userID int64) (*affiliateTransferSnapshot, error) {
	rows, queryErr := db.QueryContext(ctx, `
SELECT u.balance::double precision,
       ua.aff_quota::double precision,
       ua.aff_frozen_quota::double precision,
       ua.aff_history_quota::double precision
FROM users u
JOIN user_affiliates ua ON ua.user_id = u.id
WHERE u.id = $1
LIMIT 1`, userID)
	if queryErr != nil {
		return nil, fmt.Errorf("capture post-transfer snapshot: %w", queryErr)
	}
	defer func() { _ = rows.Close() }()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, service.ErrUserNotFound
	}

	var snap affiliateTransferSnapshot
	if scanErr := rows.Scan(
		&snap.BalanceAfter,
		&snap.AvailableQuotaAfter,
		&snap.FrozenQuotaAfter,
		&snap.HistoryQuotaAfter,
	); scanErr != nil {
		return nil, scanErr
	}
	return &snap, rows.Err()
}

func toFloat64Ptr(v sql.NullFloat64) *float64 {
	if !v.Valid {
		return nil
	}
	return &v.Float64
}

func mintAffiliateCode() (string, error) {
	raw := make([]byte, affiliateCodeLength)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("entropy read for affiliate code: %w", err)
	}
	for idx := range raw {
		raw[idx] = affiliateCodeCharset[int(raw[idx])%len(affiliateCodeCharset)]
	}
	return string(raw), nil
}

// isUniqueConstraintViolation is declared in error_translate.go; this file
// reuses that package-level function instead of redeclaring it.

// UpdateUserAffCode sets a custom affiliate code for the given user.
// Returns ErrAffiliateCodeTaken if the code collides with an existing one.
func (r *affiliateRepository) UpdateUserAffCode(ctx context.Context, userID int64, newCode string) error {
	if userID <= 0 {
		return service.ErrUserNotFound
	}
	upper := strings.ToUpper(strings.TrimSpace(newCode))
	if upper == "" {
		return service.ErrAffiliateCodeInvalid
	}

	return r.runInTx(ctx, func(txCtx context.Context, txDB *dbent.Client) error {
		if _, err := provisionUserAffiliate(txCtx, txDB, userID); err != nil {
			return err
		}
		result, execErr := txDB.ExecContext(txCtx, `
UPDATE user_affiliates
SET aff_code = $1,
    aff_code_custom = true,
    updated_at = NOW()
WHERE user_id = $2`, upper, userID)
		if execErr != nil {
			if isUniqueConstraintViolation(execErr) {
				return service.ErrAffiliateCodeTaken
			}
			return fmt.Errorf("set custom aff_code: %w", execErr)
		}
		n, _ := result.RowsAffected()
		if n == 0 {
			return service.ErrUserNotFound
		}
		return nil
	})
}

// ResetUserAffCode reverts the affiliate code to a system-generated random value
// and clears the custom flag.
func (r *affiliateRepository) ResetUserAffCode(ctx context.Context, userID int64) (string, error) {
	if userID <= 0 {
		return "", service.ErrUserNotFound
	}
	var generated string
	txErr := r.runInTx(ctx, func(txCtx context.Context, txDB *dbent.Client) error {
		if _, err := provisionUserAffiliate(txCtx, txDB, userID); err != nil {
			return err
		}
		for attempt := 0; attempt < affiliateCodeMaxAttempts; attempt++ {
			code, genErr := mintAffiliateCode()
			if genErr != nil {
				return genErr
			}
			result, execErr := txDB.ExecContext(txCtx, `
UPDATE user_affiliates
SET aff_code = $1,
    aff_code_custom = false,
    updated_at = NOW()
WHERE user_id = $2`, code, userID)
			if execErr != nil {
				if isUniqueConstraintViolation(execErr) {
					continue
				}
				return fmt.Errorf("reset aff_code: %w", execErr)
			}
			n, _ := result.RowsAffected()
			if n == 0 {
				return service.ErrUserNotFound
			}
			generated = code
			return nil
		}
		return fmt.Errorf("reset aff_code: all attempts exhausted")
	})
	if txErr != nil {
		return "", txErr
	}
	return generated, nil
}

// SetUserRebateRate sets or clears a per-user rebate rate.
// A nil ratePercent clears the override (falls back to global default).
func (r *affiliateRepository) SetUserRebateRate(ctx context.Context, userID int64, ratePercent *float64) error {
	if userID <= 0 {
		return service.ErrUserNotFound
	}
	return r.runInTx(ctx, func(txCtx context.Context, txDB *dbent.Client) error {
		if _, err := provisionUserAffiliate(txCtx, txDB, userID); err != nil {
			return err
		}
		// sqlNullableParam converts nil *float64 to SQL NULL, non-nil to its value.
		result, execErr := txDB.ExecContext(txCtx, `
UPDATE user_affiliates
SET aff_rebate_rate_percent = $1,
    updated_at = NOW()
WHERE user_id = $2`, sqlNullableParam(ratePercent), userID)
		if execErr != nil {
			return fmt.Errorf("set rebate rate: %w", execErr)
		}
		n, _ := result.RowsAffected()
		if n == 0 {
			return service.ErrUserNotFound
		}
		return nil
	})
}

// BatchSetUserRebateRate applies the same rebate rate to multiple users.
// A nil ratePercent clears per-user overrides.
func (r *affiliateRepository) BatchSetUserRebateRate(ctx context.Context, userIDs []int64, ratePercent *float64) error {
	if len(userIDs) == 0 {
		return nil
	}
	return r.runInTx(ctx, func(txCtx context.Context, txDB *dbent.Client) error {
		for _, uid := range userIDs {
			if uid <= 0 {
				continue
			}
			if _, err := provisionUserAffiliate(txCtx, txDB, uid); err != nil {
				return err
			}
		}
		_, execErr := txDB.ExecContext(txCtx, `
UPDATE user_affiliates
SET aff_rebate_rate_percent = $1,
    updated_at = NOW()
WHERE user_id = ANY($2)`, sqlNullableParam(ratePercent), pq.Array(userIDs))
		if execErr != nil {
			return fmt.Errorf("batch set rebate rate: %w", execErr)
		}
		return nil
	})
}

// sqlNullableParam unwraps a *float64 into an interface{} suitable for SQL
// parameter binding: nil pointer maps to SQL NULL, non-nil yields the float.
func sqlNullableParam(v *float64) any {
	if v == nil {
		return nil
	}
	return *v
}

func optionalInt64Param(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}

// ListUsersWithCustomSettings returns users who have a custom affiliate code or
// a per-user rebate rate.
//
// A single query handles both empty and non-empty search terms: an empty
// search produces the LIKE pattern "%%" which matches all rows; a non-empty
// search performs a case-insensitive substring match.
func (r *affiliateRepository) ListUsersWithCustomSettings(ctx context.Context, filter service.AffiliateAdminFilter) ([]service.AffiliateAdminEntry, int64, error) {
	pg := filter.Page
	if pg < 1 {
		pg = 1
	}
	pgSize := filter.PageSize
	if pgSize <= 0 || pgSize > 200 {
		pgSize = 20
	}
	off := (pg - 1) * pgSize
	pattern := "%" + strings.TrimSpace(filter.Search) + "%"

	const baseFrom = `
FROM user_affiliates ua
JOIN users u ON u.id = ua.user_id
WHERE (ua.aff_code_custom = true OR ua.aff_rebate_rate_percent IS NOT NULL)
  AND (u.email ILIKE $1 OR u.username ILIKE $1)`

	db := clientFromContext(ctx, r.client)

	total, cntErr := fetchSingleInt64(ctx, db, "SELECT COUNT(*)"+baseFrom, pattern)
	if cntErr != nil {
		return nil, 0, fmt.Errorf("count custom-settings affiliate entries: %w", cntErr)
	}

	selectSQL := `
SELECT ua.user_id,
       COALESCE(u.email, ''),
       COALESCE(u.username, ''),
       ua.aff_code,
       ua.aff_code_custom,
       ua.aff_rebate_rate_percent,
       ua.aff_count` + baseFrom + `
ORDER BY ua.updated_at DESC
LIMIT $2 OFFSET $3`

	rows, queryErr := db.QueryContext(ctx, selectSQL, pattern, pgSize, off)
	if queryErr != nil {
		return nil, 0, fmt.Errorf("list custom-settings affiliate entries: %w", queryErr)
	}
	defer func() { _ = rows.Close() }()

	entries := make([]service.AffiliateAdminEntry, 0)
	for rows.Next() {
		var e service.AffiliateAdminEntry
		var rate sql.NullFloat64
		if scanErr := rows.Scan(&e.UserID, &e.Email, &e.Username, &e.AffCode,
			&e.AffCodeCustom, &rate, &e.AffCount); scanErr != nil {
			return nil, 0, scanErr
		}
		if rate.Valid {
			v := rate.Float64
			e.AffRebateRatePercent = &v
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return entries, total, nil
}

// fetchSingleInt64 executes a query expected to return a single int64 (e.g. COUNT).
func fetchSingleInt64(ctx context.Context, db affiliateQueryRunner, query string, args ...any) (int64, error) {
	rows, queryErr := db.QueryContext(ctx, query, args...)
	if queryErr != nil {
		return 0, queryErr
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return 0, err
		}
		return 0, nil
	}
	var val int64
	if scanErr := rows.Scan(&val); scanErr != nil {
		return 0, scanErr
	}
	return val, nil
}
