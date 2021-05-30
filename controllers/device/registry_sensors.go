package device

import (
	"github.com/pkg/errors"
	"github.com/timoth-y/chainmetric-core/models/requests"
	"github.com/timoth-y/chainmetric-sensorsys/core"
	"github.com/timoth-y/chainmetric-sensorsys/core/sensor"
	"github.com/timoth-y/chainmetric-sensorsys/network/blockchain"
	"github.com/timoth-y/chainmetric-sensorsys/shared"
)

// RegisteredSensors returns map with sensors registered on the Device.
func (d *Device) RegisteredSensors() sensor.SensorsRegister {
	return d.sensors
}

// RegisterSensors adds given `sensors` on the Device sensors pool.
func (d *Device) RegisterSensors(sensors ...core.Sensor) {
	for i, s := range sensors {
		d.sensors[s.ID()] = sensors[i]
	}

	if d.IsLoggedToNetwork() {
		d.updateSupportedMetrics()
	}
}

// UnregisterSensor removes sensor by given `id` from the Device sensors pool.
func (d *Device) UnregisterSensor(id string) {
	if d.sensors.Exists(id) {
		delete(d.sensors, id)
	}

	if d.IsLoggedToNetwork() {
		d.updateSupportedMetrics()
	}
}

// UpdateSensorsRegister applies changes in sensor.SensorsRegister of the Device.
func (d *Device) UpdateSensorsRegister(added []core.Sensor, removed []string) {
	for i, s := range added {
		d.sensors[s.ID()] = added[i]
	}

	for _, id := range removed {
		delete(d.sensors, id)
	}

	if d.IsLoggedToNetwork() {
		d.updateSupportedMetrics()
	}
}

func (d *Device) updateSupportedMetrics() {
	if err := blockchain.Contracts.Devices.Update(d.ID(), requests.DeviceUpdateRequest{
		Supports: d.sensors.SupportedMetrics(),
	}); err != nil {
		shared.Logger.Error(errors.Wrap(err, "failed to update supported metrics"))
	}
}

// StaticSensors returns map with sensors statically registered on the Device.
func (d *Device) StaticSensors() sensor.SensorsRegister {
	return d.staticSensors
}

// RegisterStaticSensors allows to registrant static (not auto-detectable) sensors.
func (d *Device) RegisterStaticSensors(sensors ...core.Sensor) *Device {
	for i, s := range sensors {
		d.staticSensors[s.ID()] = sensors[i]
	}
	return d
}