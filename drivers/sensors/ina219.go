package sensors

import (
	"github.com/timoth-y/chainmetric-core/models"
	"periph.io/x/periph/experimental/devices/ina219"

	"github.com/timoth-y/chainmetric-iot/core/dev/sensor"
	"github.com/timoth-y/chainmetric-iot/drivers/periphery"
)

type INA219 struct {
	*periphery.I2C
	*ina219.Dev
}

func NewINA219(addr uint16, bus int) *INA219 {
	return &INA219{
		I2C: periphery.NewI2C(addr, bus),
	}
}

func (s *INA219) ID() string {
	return "INA219"
}

func (s *INA219) Init() (err error) {
	if err = s.I2C.Init(); err != nil {
		return
	}

	if s.Dev, err = ina219.New(s.Bus, &ina219.Opts{
		Address:       int(s.Addr),
		SenseResistor: ina219.DefaultOpts.SenseResistor,
		MaxCurrent:    ina219.DefaultOpts.MaxCurrent,
	}); err != nil {
		return
	}

	return
}

func (s *INA219) Harvest(ctx *sensor.Context) {
	if power, err := s.Sense(); err != nil {
		ctx.Error(err)
	} else {
		ctx.WriterFor("current").Write(float64(power.Current))
		ctx.WriterFor("voltage").Write(float64(power.Voltage))
	}
}

func (s *INA219) ReadVoltage() (int64, error) {
	if power, err := s.Sense(); err != nil {
		return 0, err
	} else {
		return int64(power.Voltage), nil
	}
}

func (s *INA219) ReadCurrent() (int64, error) {
	if power, err := s.Sense(); err != nil {
		return 0, err
	} else {
		return int64(power.Current), nil
	}
}

func (s *INA219) Metrics() []models.Metric {
	return []models.Metric {
		"current",
		"voltage",
	}
}

func (s *INA219) Verify() bool {
	if !s.I2C.Verify() {
		return false
	}

	if devID, err := s.I2C.ReadReg(BMP280_DEVICE_ID_REGISTER); err == nil {
		return devID == BMP280_DEVICE_ID
	}

	return false
}
