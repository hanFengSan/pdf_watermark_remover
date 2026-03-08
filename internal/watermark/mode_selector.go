package watermark

type SuppressionMode string

const (
	ModeSafe       SuppressionMode = "safe"
	ModeBalanced   SuppressionMode = "balanced"
	ModeAggressive SuppressionMode = "aggressive"
)

func SelectBestEffortMode(pattern Pattern) bool {
	return pattern.Confidence < 0.5
}

func SelectSuppressionMode(pattern Pattern, darkRatio, stdDev float64) SuppressionMode {
	if pattern.Confidence >= 0.74 && darkRatio < 0.58 && stdDev >= 18 {
		return ModeAggressive
	}
	if pattern.Confidence >= 0.52 {
		return ModeBalanced
	}
	return ModeSafe
}
