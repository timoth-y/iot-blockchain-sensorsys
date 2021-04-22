package sensors

import (
	"github.com/timoth-y/iot-blockchain-contracts/models"

	"github.com/timoth-y/iot-blockchain-sensorsys/drivers/peripherals"
	"github.com/timoth-y/iot-blockchain-sensorsys/model/metrics"
)

type AnalogHall struct {
	ch *peripherals.AnalogChannel
	active bool
}

func NewAnalogHall(ch *peripherals.AnalogChannel) *AnalogHall {
	return &AnalogHall{
		ch: ch,
	}
}

func (s *AnalogHall) ID() string {
	return "Analog_HALL"
}

func (s *AnalogHall) Init() error {
	s.active = true
	return nil
}

func (s *AnalogHall) Harvest(ctx *Context) {
	ctx.For(metrics.Magnetism).Write(s.ch.Get())
}

func (s *AnalogHall) Metrics() []models.Metric {
	return []models.Metric {
		metrics.Magnetism,
	}
}

func (s *AnalogHall) Active() bool {
	return s.active
}

func (s *AnalogHall) Close() error {
	s.active = false
	return nil
}
