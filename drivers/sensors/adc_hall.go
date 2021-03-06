package sensors

import (
	"sync"

	"github.com/spf13/viper"
	"github.com/timoth-y/chainmetric-core/models"

	"github.com/timoth-y/chainmetric-core/models/metrics"

	"github.com/timoth-y/chainmetric-iot/core/dev/sensor"
	"github.com/timoth-y/chainmetric-iot/drivers/periphery"
)

var (
	adcHallMutex = &sync.Mutex{}
)

type ADCHall struct {
	periphery.ADC
	samples int
}

func NewADCHall(addr uint16, bus int) sensor.Sensor {
	return &ADCHall{
		ADC: periphery.NewADC(addr, bus, periphery.WithConversion(func(raw float64) float64 {
			volts := raw / periphery.ADS1115_SAMPLES_PER_READ * periphery.ADS1115_VOLTS_PER_SAMPLE
			return volts * 1000 / ADC_HALL_SENSITIVITY
		}), periphery.WithBias(ADC_HALL_BIAS), periphery.WithI2CMutex(adcHallMutex)),
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
	ctx.WriterFor(metrics.Magnetism).Write(s.Read())
}

func (s *ADCHall) Metrics() []models.Metric {
	return []models.Metric {
		metrics.Magnetism,
	}
}
