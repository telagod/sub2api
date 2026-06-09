package service

// resolveImageRateMultiplier determines the effective image billing multiplier.
// When the group has an independent image rate setting, that value is used
// (clamped to zero if negative). Otherwise the inherited group multiplier applies.
func resolveImageRateMultiplier(key *APIKey, groupMultiplier float64) float64 {
	if key == nil || key.Group == nil || !key.Group.ImageRateIndependent {
		return groupMultiplier
	}
	mult := key.Group.ImageRateMultiplier
	if mult < 0 {
		mult = 0
	}
	return mult
}
