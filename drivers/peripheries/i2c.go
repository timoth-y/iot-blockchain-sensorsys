package peripheries

import (
	"fmt"

	"github.com/pkg/errors"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"

	"github.com/timoth-y/chainmetric-sensorsys/shared"
)

// I2C provides wrapper for I2C peripheral
type I2C struct {
	i2c.Dev
	name   string
	bus    i2c.BusCloser
	active bool
}

// NewI2C creates new I2C driver instance.
func NewI2C(addr uint16, bus int) *I2C {
	return &I2C{
		Dev: i2c.Dev{
			Addr: addr,
		},
		name: shared.NtoI2cBusName(bus),
	}
}

// Init performs I2C device initialization.
func (i *I2C) Init() (err error) {
	if i.bus, err = i2creg.Open(i.name); err != nil {
		return errors.Wrapf(err, "failed to open an I2C bus on %s", i.name)
	}

	i.Bus = i.bus
	i.active = true

	return
}

// Read reads a single byte from an active register.
func (i *I2C) Read() (byte, error) {
	b := make([]byte, 1)
	if err := i.Tx(nil, b); err != nil {
		return 0, err
	}

	return b[0], nil
}

// ReadReg reads `n` bytes from an active register.
func (i *I2C) ReadBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if err := i.Tx(nil, b); err != nil {
		return nil, err
	}

	return b, nil
}

// ReadReg reads a single byte from a specified `reg` register.
func (i *I2C) ReadReg(reg byte) (byte, error) {
	b := make([]byte, 1)
	if err := i.Tx([]byte{reg}, b); err != nil {
		return 0, err
	}

	return b[0], nil
}

// ReadRegBytes reads `n` bytes from a specified `reg` register.
func (i *I2C) ReadRegBytes(reg byte, n int) ([]byte, error) {
	b := make([]byte, n)
	if err := i.Tx([]byte{reg}, b); err != nil {
		return nil, err
	}

	return b, nil
}

// ReadRegU16BE reads unsigned big endian word (16 bits) from a specified `reg` register.
func (i *I2C) ReadRegU16BE(reg byte) (uint16, error) {
	buf, err := i.ReadRegBytes(reg, 2)
	if err != nil {
		return 0, err
	}

	return uint16(buf[0])<<8 + uint16(buf[1]), nil
}

// ReadRegU16LE reads unsigned little endian word (16 bits) from a specified `reg` register.
func (i *I2C) ReadRegU16LE(reg byte) (uint16, error) {
	w, err := i.ReadRegU16BE(reg)
	if err != nil {
		return 0, err
	}

	// exchange bytes
	return (w&0xFF)<<8 + w>>8, nil
}

// ReadRegS16BE reads signed big endian word (16 bits) from a specified `reg` register.
func (i *I2C) ReadRegS16BE(reg byte) (int16, error) {
	buf, err := i.ReadRegBytes(reg, 2)
	if err != nil {
		return 0, err
	}

	return int16(buf[0])<<8 + int16(buf[1]), nil
}

// ReadRegS16LE reads signed little endian word (16 bits) from a specified `reg` register.
func (i *I2C) ReadRegS16LE(reg byte) (int16, error) {
	w, err := i.ReadRegS16BE(reg)
	if err != nil {
		return 0, err
	}

	// exchange bytes
	return (w&0xFF)<<8 + w>>8, nil

}

// WriteBytes writes `data` bytes to an active register.
func (i *I2C) WriteBytes(data ...byte) error {
	n, err := i.Write(data)
	if err != nil {
		return err
	}

	if n != len(data) {
		return fmt.Errorf("write: wrong number of bytes written: want %d, got %d", len(data), n)
	}

	return nil
}

// WriteRegBytes writes `data` bytes to a specified `reg` register.
func (i *I2C) WriteRegBytes(reg byte, data ...byte) error {
	n, err := i.Write(append([]byte{reg}, data...))
	if err != nil {
		return err
	}

	if n - 1 != len(data) {
		return fmt.Errorf("write: wrong number of bytes written: want %d, got %d", len(data), n - 1)
	}

	return nil
}

// Tx wraps i2c.Dev Tx() method with activeness check.
func (i *I2C) Tx(w, r []byte) error {
	if i.active {
		return i.Dev.Tx(w, r)
	}

	return nil
}

// Verify verifies I2C bus connectivity.
// It will perform Init if driver is not Active.
func (i *I2C) Verify() bool {
	if !i.active {
		if err := i.Init(); err != nil {
			return false
		}
	}

	return true
}

// Active checks whether the I2C device is connected and active.
func (i *I2C) Active() bool {
	return i.active
}

// Close closes connection to I2C device and clears allocated resources.
func (i *I2C) Close() error {
	i.active = false
	return i.bus.Close()
}