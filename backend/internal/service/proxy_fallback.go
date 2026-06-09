package service

import "time"

// ResolveProxyFallbackTarget determines which proxy (or direct connection) an
// expired proxy should fall back to.
//
// Returns (targetID, changed):
//   - changed=false: no action needed (mode is none, or chain is cyclic/broken)
//   - changed=true, targetID=nil: switch to direct connection
//   - changed=true, targetID!=nil: switch to the returned backup proxy
func ResolveProxyFallbackTarget(origin Proxy, all map[int64]Proxy, now time.Time) (*int64, bool) {
	if origin.FallbackMode == FallbackModeDirect {
		return nil, true
	}
	if origin.FallbackMode != FallbackModeProxy {
		return nil, false
	}

	seen := map[int64]bool{origin.ID: true}
	nextID := origin.BackupProxyID

	for nextID != nil {
		nid := *nextID
		if seen[nid] {
			return nil, false
		}
		candidate, exists := all[nid]
		if !exists {
			return nil, false
		}
		if !(&candidate).IsExpired(now) && candidate.Status != StatusExpired {
			result := candidate.ID
			return &result, true
		}
		seen[nid] = true

		switch candidate.FallbackMode {
		case FallbackModeDirect:
			return nil, true
		case FallbackModeProxy:
			nextID = candidate.BackupProxyID
		default:
			return nil, false
		}
	}
	return nil, false
}
