package sensors

import (
	"github.com/timoth-y/chainmetric-core/models"

	"github.com/timoth-y/chainmetric-core/models/metrics"

	"github.com/timoth-y/chainmetric-sensorsys/drivers/peripherals"
	"github.com/timoth-y/chainmetric-sensorsys/drivers/sensor"
)

type ADCPiezo struct {
	peripherals.ADC
}

func NewADCPiezo(addr uint16, bus int) sensor.Sensor {
	return &ADCPiezo{
		ADC: peripherals.NewADC(addr, bus),
	}
}

func (s *ADCPiezo) ID() string {
	return "ADC_Piezo"
}

func (s *ADCPiezo) Read() float64 {
	return s.Aggregate(100, nil) - ADC_PIEZO_BIAS
}

func (s *ADCPiezo) Harvest(ctx *sensor.Context) {
	ctx.For(metrics.Vibration).WriteWithError(s.ReadRetry(5))
}

func (s *ADCPiezo) Metrics() []models.Metric {
	return []models.Metric {
		metrics.Vibration,
	}
}
