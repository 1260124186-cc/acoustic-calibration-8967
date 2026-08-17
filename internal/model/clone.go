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

func CloneRun(source CalibrationRun) CalibrationRun {
	return source
}
