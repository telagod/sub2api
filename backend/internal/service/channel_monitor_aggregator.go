package service

import (
	"context"
	"fmt"
	"log/slog"
)

// Channel monitor aggregation layer: combines latest check state and availability
// metrics into summary/detail views for admin and user interfaces.
// All methods follow the "fail gracefully, log only" principle to prevent N+1
// query failures from blocking list rendering.

// BatchMonitorStatusSummary aggregates latest status and 7-day availability for
// multiple monitors in batch (eliminates N+1 for admin/user list views).
// On failure, returns an empty map and logs a warning.
//
// Parameters:
//   - monitorIDs: IDs to aggregate
//   - primaryByID: monitor ID -> primary model name (for 7d availability and latest status)
//   - extrasByID: monitor ID -> extra model list (for latest status in ExtraModels)
func (s *ChannelMonitorService) BatchMonitorStatusSummary(
	reqCtx context.Context,
	monitorIDs []int64,
	primaryByID map[int64]string,
	extrasByID map[int64][]string,
) map[int64]MonitorStatusSummary {
	result := make(map[int64]MonitorStatusSummary, len(monitorIDs))
	if len(monitorIDs) == 0 {
		return result
	}
	latestRows, latestErr := s.repo.ListLatestForMonitorIDs(reqCtx, monitorIDs)
	if latestErr != nil {
		slog.Warn("channel_monitor: failed to batch-load latest records", "error", latestErr)
		latestRows = map[int64][]*ChannelMonitorLatest{}
	}
	availRows, availErr := s.repo.ComputeAvailabilityForMonitors(reqCtx, monitorIDs, monitorAvailability7Days)
	if availErr != nil {
		slog.Warn("channel_monitor: failed to batch-compute availability", "error", availErr)
		availRows = map[int64][]*ChannelMonitorAvailability{}
	}

	for idx := 0; idx < len(monitorIDs); idx++ {
		mid := monitorIDs[idx]
		result[mid] = buildStatusSummary(
			indexLatestByModel(latestRows[mid]),
			indexAvailabilityByModel(availRows[mid]),
			primaryByID[mid],
			extrasByID[mid],
		)
	}
	return result
}

// ListUserView returns the read-only user overview of all enabled monitors.
// Uses batch queries to avoid N+1:
//   - 1 query for monitors
//   - 1 batch query for latest (including ping_latency_ms)
//   - 1 batch query for 7d availability
//   - 1 batch query for timeline (primary model recent N entries)
func (s *ChannelMonitorService) ListUserView(reqCtx context.Context) ([]*UserMonitorView, error) {
	allMonitors, fetchErr := s.repo.ListEnabled(reqCtx)
	if fetchErr != nil {
		return nil, fmt.Errorf("list enabled monitors: %w", fetchErr)
	}
	if len(allMonitors) == 0 {
		return []*UserMonitorView{}, nil
	}

	monitorIDs, primaryByID, extrasByID := collectMonitorIndexes(allMonitors)
	summaryMap := s.BatchMonitorStatusSummary(reqCtx, monitorIDs, primaryByID, extrasByID)
	latestRows := s.batchLatest(reqCtx, monitorIDs)
	timelineRows := s.batchTimeline(reqCtx, monitorIDs, primaryByID)

	viewList := make([]*UserMonitorView, 0, len(allMonitors))
	for idx := 0; idx < len(allMonitors); idx++ {
		mon := allMonitors[idx]
		primLatest := pickLatest(latestRows[mon.ID], mon.PrimaryModel)
		viewList = append(viewList, buildUserViewFromSummary(mon, summaryMap[mon.ID], primLatest, timelineRows[mon.ID]))
	}
	return viewList, nil
}

// collectMonitorIndexes extracts the three index structures needed for batch aggregation.
func collectMonitorIndexes(monitors []*ChannelMonitor) ([]int64, map[int64]string, map[int64][]string) {
	monitorIDs := make([]int64, 0, len(monitors))
	primaryByID := make(map[int64]string, len(monitors))
	extrasByID := make(map[int64][]string, len(monitors))
	for idx := 0; idx < len(monitors); idx++ {
		mon := monitors[idx]
		monitorIDs = append(monitorIDs, mon.ID)
		primaryByID[mon.ID] = mon.PrimaryModel
		extrasByID[mon.ID] = mon.ExtraModels
	}
	return monitorIDs, primaryByID, extrasByID
}

// batchLatest fetches latest status per model in batch; logs on failure without blocking rendering.
func (s *ChannelMonitorService) batchLatest(reqCtx context.Context, monitorIDs []int64) map[int64][]*ChannelMonitorLatest {
	rows, fetchErr := s.repo.ListLatestForMonitorIDs(reqCtx, monitorIDs)
	if fetchErr != nil {
		slog.Warn("channel_monitor: user view batch latest query failed", "error", fetchErr)
		return map[int64][]*ChannelMonitorLatest{}
	}
	return rows
}

// batchTimeline fetches the most recent monitorTimelineMaxPoints history entries
// for each monitor's primary model.
func (s *ChannelMonitorService) batchTimeline(
	reqCtx context.Context,
	monitorIDs []int64,
	primaryByID map[int64]string,
) map[int64][]*ChannelMonitorHistoryEntry {
	rows, fetchErr := s.repo.ListRecentHistoryForMonitors(reqCtx, monitorIDs, primaryByID, monitorTimelineMaxPoints)
	if fetchErr != nil {
		slog.Warn("channel_monitor: user view batch timeline query failed", "error", fetchErr)
		return map[int64][]*ChannelMonitorHistoryEntry{}
	}
	return rows
}

// pickLatest selects the latest entry matching the given model from a slice; returns nil if not found.
func pickLatest(entries []*ChannelMonitorLatest, modelName string) *ChannelMonitorLatest {
	if modelName == "" {
		return nil
	}
	for idx := 0; idx < len(entries); idx++ {
		if entries[idx].Model == modelName {
			return entries[idx]
		}
	}
	return nil
}

// GetUserDetail returns a single monitor's read-only detail view with per-model
// availability across 7/15/30 day windows. API key is not exposed.
func (s *ChannelMonitorService) GetUserDetail(reqCtx context.Context, monitorID int64) (*UserMonitorDetail, error) {
	mon, fetchErr := s.repo.GetByID(reqCtx, monitorID)
	if fetchErr != nil {
		return nil, fetchErr
	}
	if !mon.Enabled {
		return nil, ErrChannelMonitorNotFound
	}

	latestEntries, latestErr := s.repo.ListLatestPerModel(reqCtx, monitorID)
	if latestErr != nil {
		return nil, fmt.Errorf("list latest per model: %w", latestErr)
	}
	windowMap, windowErr := s.collectAvailabilityWindows(reqCtx, monitorID)
	if windowErr != nil {
		return nil, windowErr
	}

	modelDetails := mergeModelDetails(mon, latestEntries, windowMap)
	return &UserMonitorDetail{
		ID:        mon.ID,
		Name:      mon.Name,
		Provider:  mon.Provider,
		GroupName: mon.GroupName,
		Models:    modelDetails,
	}, nil
}

// collectAvailabilityWindows queries the 7/15/30 day windows in one pass, indexed by model.
func (s *ChannelMonitorService) collectAvailabilityWindows(reqCtx context.Context, monitorID int64) (map[int]map[string]*ChannelMonitorAvailability, error) {
	windowMap := make(map[int]map[string]*ChannelMonitorAvailability, 3)
	dayWindows := [3]int{monitorAvailability7Days, monitorAvailability15Days, monitorAvailability30Days}
	for _, days := range dayWindows {
		entries, queryErr := s.repo.ComputeAvailability(reqCtx, monitorID, days)
		if queryErr != nil {
			return nil, fmt.Errorf("compute availability %dd: %w", days, queryErr)
		}
		windowMap[days] = indexAvailabilityByModel(entries)
	}
	return windowMap, nil
}

// ---------- Pure helper functions (no IO, reusable across batch/single/detail paths) ----------

// indexLatestByModel builds a model->latest lookup map from a slice.
func indexLatestByModel(entries []*ChannelMonitorLatest) map[string]*ChannelMonitorLatest {
	lookup := make(map[string]*ChannelMonitorLatest, len(entries))
	for idx := 0; idx < len(entries); idx++ {
		lookup[entries[idx].Model] = entries[idx]
	}
	return lookup
}

// indexAvailabilityByModel builds a model->availability lookup map from a slice.
func indexAvailabilityByModel(entries []*ChannelMonitorAvailability) map[string]*ChannelMonitorAvailability {
	lookup := make(map[string]*ChannelMonitorAvailability, len(entries))
	for idx := 0; idx < len(entries); idx++ {
		lookup[entries[idx].Model] = entries[idx]
	}
	return lookup
}

// buildStatusSummary assembles a MonitorStatusSummary from latest and availability maps.
// Pure assembly with no IO; suitable for both batch and single-monitor paths.
func buildStatusSummary(
	latestByModel map[string]*ChannelMonitorLatest,
	availByModel map[string]*ChannelMonitorAvailability,
	primaryModel string,
	extraModels []string,
) MonitorStatusSummary {
	out := MonitorStatusSummary{ExtraModels: make([]ExtraModelStatus, 0, len(extraModels))}
	if primaryModel != "" {
		if latestEntry, found := latestByModel[primaryModel]; found {
			out.PrimaryStatus = latestEntry.Status
			out.PrimaryLatencyMs = latestEntry.LatencyMs
		}
		if availEntry, found := availByModel[primaryModel]; found {
			out.Availability7d = availEntry.AvailabilityPct
		}
	}
	for idx := 0; idx < len(extraModels); idx++ {
		modelName := extraModels[idx]
		extraEntry := ExtraModelStatus{Model: modelName}
		if latestEntry, found := latestByModel[modelName]; found {
			extraEntry.Status = latestEntry.Status
			extraEntry.LatencyMs = latestEntry.LatencyMs
		}
		out.ExtraModels = append(out.ExtraModels, extraEntry)
	}
	return out
}

// buildUserViewFromSummary populates a UserMonitorView from pre-aggregated summary,
// primary model latest, and timeline entries (no IO).
// primaryLatest may be nil (no history yet); timelineEntries may be empty.
func buildUserViewFromSummary(
	mon *ChannelMonitor,
	summary MonitorStatusSummary,
	primaryLatest *ChannelMonitorLatest,
	timelineEntries []*ChannelMonitorHistoryEntry,
) *UserMonitorView {
	uv := &UserMonitorView{
		ID:               mon.ID,
		Name:             mon.Name,
		Provider:         mon.Provider,
		GroupName:        mon.GroupName,
		PrimaryModel:     mon.PrimaryModel,
		PrimaryStatus:    summary.PrimaryStatus,
		PrimaryLatencyMs: summary.PrimaryLatencyMs,
		Availability7d:   summary.Availability7d,
		ExtraModels:      summary.ExtraModels,
		Timeline:         buildTimelinePoints(timelineEntries),
	}
	if primaryLatest != nil {
		uv.PrimaryPingLatencyMs = primaryLatest.PingLatencyMs
	}
	return uv
}

// buildTimelinePoints converts history entries to lightweight timeline points
// (strips message/ID/Model to reduce response payload).
func buildTimelinePoints(entries []*ChannelMonitorHistoryEntry) []UserMonitorTimelinePoint {
	points := make([]UserMonitorTimelinePoint, 0, len(entries))
	for idx := 0; idx < len(entries); idx++ {
		entry := entries[idx]
		points = append(points, UserMonitorTimelinePoint{
			Status:        entry.Status,
			LatencyMs:     entry.LatencyMs,
			PingLatencyMs: entry.PingLatencyMs,
			CheckedAt:     entry.CheckedAt,
		})
	}
	return points
}

// mergeModelDetails combines latest status and multi-window availability into ModelDetail list.
// Reuses indexLatestByModel to avoid duplicate map construction.
func mergeModelDetails(
	mon *ChannelMonitor,
	latestEntries []*ChannelMonitorLatest,
	windowMap map[int]map[string]*ChannelMonitorAvailability,
) []ModelDetail {
	allModels := make([]string, 0, 1+len(mon.ExtraModels))
	allModels = append(allModels, mon.PrimaryModel)
	allModels = append(allModels, mon.ExtraModels...)
	latestLookup := indexLatestByModel(latestEntries)
	details := make([]ModelDetail, 0, len(allModels))
	for idx := 0; idx < len(allModels); idx++ {
		modelName := allModels[idx]
		detail := ModelDetail{Model: modelName}
		if latestEntry, found := latestLookup[modelName]; found {
			detail.LatestStatus = latestEntry.Status
			detail.LatestLatencyMs = latestEntry.LatencyMs
		}
		if avail7, found := windowMap[monitorAvailability7Days][modelName]; found {
			detail.Availability7d = avail7.AvailabilityPct
			detail.AvgLatency7dMs = avail7.AvgLatencyMs
		}
		if avail15, found := windowMap[monitorAvailability15Days][modelName]; found {
			detail.Availability15d = avail15.AvailabilityPct
		}
		if avail30, found := windowMap[monitorAvailability30Days][modelName]; found {
			detail.Availability30d = avail30.AvailabilityPct
		}
		details = append(details, detail)
	}
	return details
}
