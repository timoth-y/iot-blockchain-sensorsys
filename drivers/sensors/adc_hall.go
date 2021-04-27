package sensors

import (
	"github.com/spf13/viper"
	"github.com/timoth-y/chainmetric-core/models"

	"github.com/timoth-y/chainmetric-core/models/metrics"

	"github.com/timoth-y/chainmetric-sensorsys/drivers/peripherals"
	"github.com/timoth-y/chainmetric-sensorsys/drivers/sensor"
)

type ADCHall struct {
	peripherals.ADC
	samples int
}

func NewADCHall(addr uint16, bus int) sensor.Sensor {
	return &ADCHall{
		ADC: peripherals.NewADC(addr, bus, peripherals.WithConversion(func(raw float64) float64 {
			volts := raw / peripherals.ADS1115_SAMPLES_PER_READ * peripherals.ADS1115_VOLTS_PER_SAMPLE
			return volts * 1000 / ADC_HALL_SENSITIVITY
		}), peripherals.WithBias(ADC_HALL_BIAS)),
		samples: viper.GetInt("sensors.analog.samples_per_read"),
	}
}

func (s *ADCHall) ID() string {
	return "ADC_Hall"
}

func (s *ADCHall) Read() float64 {
	return s.RMS(s.samples, nil)
}

func (s *ADCHall) Harvest(ctx *sensor.Context) {
	ctx.For(metrics.Magnetism).Write(s.Read())
}

func (s *ADCHall) Metrics() []models.Metric {
	return []models.Metric {
		metrics.Magnetism,
	}
}
