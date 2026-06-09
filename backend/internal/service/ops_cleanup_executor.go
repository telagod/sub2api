package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const (
	opsCleanupDefaultSchedule  = "0 2 * * *"
	opsCleanupBatchSize        = 5000
	opsCleanupCronStopTimeout  = 3 * time.Second
	opsCleanupRunTimeout       = 30 * time.Minute
	opsCleanupHeartbeatTimeout = 2 * time.Second
)

type opsCleanupTarget struct {
	retentionDays int
	table         string
	timeCol       string
	castDate      bool
	counter       *int64
}

type opsCleanupDeletedCounts struct {
	errorLogs     int64
	alertEvents   int64
	systemLogs    int64
	logAudits     int64
	systemMetrics int64
	hourlyPreagg  int64
	dailyPreagg   int64
}

func (dc opsCleanupDeletedCounts) String() string {
	return fmt.Sprintf(
		"error_logs=%d alert_events=%d system_logs=%d log_audits=%d system_metrics=%d hourly_preagg=%d daily_preagg=%d",
		dc.errorLogs,
		dc.alertEvents,
		dc.systemLogs,
		dc.logAudits,
		dc.systemMetrics,
		dc.hourlyPreagg,
		dc.dailyPreagg,
	)
}

// opsCleanupPlan translates a retention-day count into a cleanup action.
//   - days < 0  => skip (ok=false), keep backward-compatible data
//   - days == 0 => TRUNCATE TABLE (O(1) full wipe), truncate=true
//   - days > 0  => batched DELETE for rows older than now-N days, cutoff = now - N days
func opsCleanupPlan(now time.Time, days int) (cutoff time.Time, truncate, ok bool) {
	if days < 0 {
		return time.Time{}, false, false
	}
	if days == 0 {
		return time.Time{}, true, true
	}
	return now.AddDate(0, 0, -days), false, true
}

func opsCleanupRunOne(
	ctx context.Context,
	conn *sql.DB,
	doTruncate bool,
	threshold time.Time,
	tableName, timeColumn string,
	useDateCast bool,
	chunkSize int,
) (int64, error) {
	if doTruncate {
		return truncateOpsTable(ctx, conn, tableName)
	}
	return deleteOldRowsByID(ctx, conn, tableName, timeColumn, threshold, chunkSize, useDateCast)
}

func deleteOldRowsByID(
	ctx context.Context,
	conn *sql.DB,
	tableName string,
	colName string,
	threshold time.Time,
	chunkSize int,
	useDateCast bool,
) (int64, error) {
	if conn == nil {
		return 0, nil
	}
	if chunkSize <= 0 {
		chunkSize = opsCleanupBatchSize
	}

	condition := fmt.Sprintf("%s < $1", colName)
	if useDateCast {
		condition = fmt.Sprintf("%s < $1::date", colName)
	}

	stmt := fmt.Sprintf(`
WITH batch AS (
  SELECT id FROM %s
  WHERE %s
  ORDER BY id
  LIMIT $2
)
DELETE FROM %s
WHERE id IN (SELECT id FROM batch)
`, tableName, condition, tableName)

	var removed int64
	for {
		execResult, execErr := conn.ExecContext(ctx, stmt, threshold, chunkSize)
		if execErr != nil {
			if isMissingRelationError(execErr) {
				return removed, nil
			}
			return removed, execErr
		}
		rowCount, countErr := execResult.RowsAffected()
		if countErr != nil {
			return removed, countErr
		}
		removed += rowCount
		if rowCount == 0 {
			break
		}
	}
	return removed, nil
}

// truncateOpsTable wipes a table using TRUNCATE TABLE, returning the pre-truncate row count for logging.
func truncateOpsTable(ctx context.Context, conn *sql.DB, tableName string) (int64, error) {
	if conn == nil {
		return 0, nil
	}
	var rowCount int64
	scanErr := conn.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&rowCount)
	if scanErr != nil {
		if isMissingRelationError(scanErr) {
			return 0, nil
		}
		return 0, fmt.Errorf("count %s: %w", tableName, scanErr)
	}
	if rowCount == 0 {
		return 0, nil
	}
	_, execErr := conn.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s", tableName))
	if execErr != nil {
		if isMissingRelationError(execErr) {
			return 0, nil
		}
		return 0, fmt.Errorf("truncate %s: %w", tableName, execErr)
	}
	return rowCount, nil
}

func isMissingRelationError(err error) bool {
	if err == nil {
		return false
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "relation") && strings.Contains(lower, "does not exist")
}
