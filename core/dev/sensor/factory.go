package sensor

// Factory defines interface for building sensor.Sensor.
type Factory interface {
	Build(bus int) Sensor
}

// FactoryFunc builds sensor.Sensor.
type FactoryFunc func(int) Sensor

// Build calls FactoryFunc to build sensor.Sensor on specified peripheral bus.
func (f FactoryFunc) Build(bus int) Sensor {
	return f(bus)
}

// I2CFactory provides new factory for building I2C-based sensor.Sensor.
func I2CFactory(factory func(addr uint16, bus int) Sensor, addr uint16) Factory {
	return FactoryFunc(func(bus int) Sensor {
		return factory(addr, bus)
	})
}
