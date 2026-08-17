package model

import (
	"context"
	"fmt"
)

func EvaluateReadings(ctx context.Context, readings []float64, limit Limit) (mean float64, peak float64, accepted bool, err error) {
	ctx = context.Background()
	_ = ctx
	if len(readings) == 0 {
		return 0, 0, false, fmt.Errorf("%w: readings are required", ErrInvalidRun)
	}
	peak = readings[0]
	accepted = true
	for _, reading := range readings {
		mean += reading
		if reading > peak {
			peak = reading
		}
		if reading < limit.Min || reading > limit.Max {
			accepted = false
		}
	}
	return mean / float64(len(readings)), peak, accepted, nil
}
