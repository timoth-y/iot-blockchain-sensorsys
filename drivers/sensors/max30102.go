package sensors

import (
	"github.com/cgxeiji/max3010x"
	"github.com/timoth-y/chainmetric-core/models"

	"github.com/timoth-y/chainmetric-sensorsys/drivers/sensor"
	"github.com/timoth-y/chainmetric-sensorsys/model/metrics"
	"github.com/timoth-y/chainmetric-sensorsys/shared"
)

type MAX30102 struct {
	*max3010x.Device
	addr uint16
	bus int
}

func NewMAX30102(addr uint16, bus int) sensor.Sensor {
	return &MAX30102{
		addr: addr,
		bus: bus,
	}
}

func (s *MAX30102) ID() string {
	return "MAX30102"
}

func (s *MAX30102) Init() (err error) {
	s.Device, err = max3010x.New(
		max3010x.WithSpecificBus(shared.NtoI2cBusName(s.bus)),
		max3010x.WithAddress(s.addr),
	); if err != nil {
		return
	}

	if err = s.Startup(); err != nil {
		return err
	}

	return
}

func (s *MAX30102) Harvest(ctx *sensor.Context) {
	ctx.For(metrics.HeartRate).WriteWithError(s.HeartRate())
	ctx.For(metrics.BloodOxidation).WriteWithError(s.SpO2())
}

func (s *MAX30102) Metrics() []models.Metric {
	return []models.Metric {
		metrics.HeartRate,
		metrics.BloodOxidation,
	}
}

func (s *MAX30102) Active() bool {
	return s.Device != nil
}

// Close disconnects from the device
func (s *MAX30102) Close() error {
	s.Device.Close()
	s.Device = nil
	return nil
}
