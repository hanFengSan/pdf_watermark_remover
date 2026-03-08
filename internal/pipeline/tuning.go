package pipeline

type TuningConfig struct {
	BaseWindowSize         int
	BaseSensitivity        float64
	RetryWindowSize        int
	RetrySensitivity       float64
	RetryThresholdBoost    int
	RetryThresholdMax      int
	FallbackMinScore       float64
	FallbackMinFragments   int
	FallbackMinImprovement float64
}

func DefaultTuningConfig() TuningConfig {
	return TuningConfig{
		BaseWindowSize:         31,
		BaseSensitivity:        0.10,
		RetryWindowSize:        35,
		RetrySensitivity:       0.14,
		RetryThresholdBoost:    8,
		RetryThresholdMax:      235,
		FallbackMinScore:       0.95,
		FallbackMinFragments:   14,
		FallbackMinImprovement: 0.08,
	}
}
