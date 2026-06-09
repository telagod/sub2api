package service

import "strings"

func normalizeGroupModelsListConfig(input GroupModelsListConfig) GroupModelsListConfig {
	result := GroupModelsListConfig{Enabled: input.Enabled}
	if len(input.Models) == 0 {
		return result
	}

	dedup := make(map[string]struct{}, len(input.Models))
	filtered := make([]string, 0, len(input.Models))
	for _, m := range input.Models {
		name := strings.TrimSpace(m)
		if name == "" {
			continue
		}
		if _, dup := dedup[name]; dup {
			continue
		}
		dedup[name] = struct{}{}
		filtered = append(filtered, name)
	}
	if len(filtered) > 0 {
		result.Models = filtered
	}
	return result
}

func (g *Group) CustomModelsListEnabled() bool {
	if g == nil {
		return false
	}
	return g.ModelsListConfig.Enabled && len(g.ModelsListConfig.Models) > 0
}
