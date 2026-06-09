package service

// containsInt64 reports whether target appears in the given slice.
func containsInt64(values []int64, target int64) bool {
	for idx := 0; idx < len(values); idx++ {
		if values[idx] == target {
			return true
		}
	}
	return false
}
