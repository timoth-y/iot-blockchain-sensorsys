package sensor

import (
	"github.com/timoth-y/chainmetric-core/models"
)

// ReadingResult defines structure for storing readings result from a single sensor.Sensor device.
type ReadingResult struct {
	Source string
	Value float64
}

// ReadingsPipe maps where to dump sensor.Sensor ReadingResult for concrete models.Metric.
type ReadingsPipe map[models.Metric] chan ReadingResult
