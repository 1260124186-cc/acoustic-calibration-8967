package model

func DefaultLimits() map[string]Limit {
	return map[string]Limit{"reading": {Min: 0, Max: 120}}
}

func CloneLimits(source map[string]Limit) map[string]Limit {
	if source == nil {
		return nil
	}
	copyOf := make(map[string]Limit, len(source))
	for key, value := range source {
		copyOf[key] = value
	}
	return copyOf
}

func CloneInstrument(source Instrument) Instrument {
	source.Limits = CloneLimits(source.Limits)
	return source
}

// CloneReadings 复制读数切片，避免与原切片共享底层数组
func CloneReadings(source []float64) []float64 {
	if source == nil {
		return nil
	}
	copyOf := make([]float64, len(source))
	copy(copyOf, source)
	return copyOf
}

func CloneRun(source CalibrationRun) CalibrationRun {
	source.Readings = CloneReadings(source.Readings)
	return source
}
