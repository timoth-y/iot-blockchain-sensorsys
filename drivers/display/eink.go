package display

import (
	"image"
	"image/color"
	"image/draw"
	"time"

	"github.com/pkg/errors"
	"github.com/timoth-y/chainmetric-iot/core/dev"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/devices/ssd1306/image1bit"

	"github.com/timoth-y/chainmetric-iot/drivers/periphery"
	"github.com/timoth-y/chainmetric-iot/model/config"
)

// EInk is an implementation of dev.Display driver for E-Ink 2.13" display.
type EInk struct {
	*periphery.SPI

	dc   *periphery.GPIO
	cs   *periphery.GPIO
	rst  *periphery.GPIO
	busy *periphery.GPIO

	rect image.Rectangle

	config config.DisplayConfig
}

// NewEInk creates new EInk driver instance by implementing dev.Display interface.
func NewEInk(config config.DisplayConfig) dev.Display {
	return &EInk{
		SPI:    periphery.NewSPI(config.Bus),
		dc:     periphery.NewGPIO(config.DCPin),
		cs:     periphery.NewGPIO(config.CSPin),
		rst:    periphery.NewGPIO(config.ResetPin),
		busy:   periphery.NewGPIO(config.BusyPin),
		rect:   image.Rect(0, 0, config.Height, config.Width),
		config: config,
	}
}

// Init performs EInk display device initialization.
func (d *EInk) Init() error {
	if err := d.dc.Init(); err != nil {
		return errors.Wrap(err, "error during connecting to EInk display DC pin")
	}

	if err := d.cs.Init(); err != nil {
		return errors.Wrap(err, "error during connecting to EInk display CS pin")
	}

	if err := d.rst.Init(); err != nil {
		return errors.Wrap(err, "error during connecting to EInk display RST pin")
	}

	if err := d.busy.Init(); err != nil {
		return errors.Wrap(err, "error during connecting to EInk display BUSY pin")
	}

	if err := d.SPI.Init(); err != nil {
		return errors.Wrap(err, "error during connecting to EInk display via SPI")
	}

	if err := d.init(); err != nil {
		return errors.Wrap(err, "error during initialising to EInk display driver")
	}

	return d.Clear()
}

// DrawRaw implements display.Drawer.
func (d *EInk) DrawRaw(r image.Rectangle, src image.Image, sp image.Point) error {
	var (
		xStart = sp.X
		yStart = sp.Y
		bitMap = image1bit.NewVerticalLSB(src.Bounds())
		byteToSend byte = 0x00
	)

	draw.Src.Draw(bitMap, r, src, sp)
	for i := 0; i < 3; i++ {
		bitMap = rotateBitMap(bitMap)
	}

	xEnd := xStart + bitMap.Rect.Dx() - 1
	if xStart + bitMap.Rect.Dx() >= d.rect.Dx() {
		xEnd = bitMap.Rect.Dx() - 1
	}

	yEnd := yStart + bitMap.Rect.Dy() - 1
	if yStart + bitMap.Rect.Dy() >= d.rect.Dy() {
		yEnd = bitMap.Rect.Dy() - 1
	}

	if err := d.setMemoryArea(xStart, yStart, xEnd, yEnd); err != nil {
		return err
	}

	for y := yStart; y < yEnd+1; y++ {
		if err := d.setMemoryPointer(xStart, y); err != nil {
			return err
		}
		if err := d.SendCommandArgs(writeRAM); err != nil {
			return err
		}
		for x := xStart; x < xEnd+1; x++ {
			bit := bitMap.BitAt(x - xStart, y - yStart)

			if bit {
				byteToSend |= 0x80 >> (uint32(x) % 8)
			}

			if x%8 == 7 {
				if err := d.SendData(byteToSend); err != nil {
					return err
				}
				byteToSend = 0x00
			}
		}
	}

	return nil
}

// Draw sends `src` image binary representation to EInk display buffer.
// Use Refresh() or DrawAndRefresh() to display image.
func (d *EInk) Draw(src image.Image) error {
	return d.DrawRaw(src.Bounds(), src, image.Point{})
}

// DrawAndRefresh sends `src` image binary representation to EInk display buffer
// and triggers update of the frame.
func (d *EInk) DrawAndRefresh(src image.Image) error {
	if err := d.Draw(src); err != nil {
		return err
	}

	return d.Refresh()
}

// ResetFrameMemory clear the frame memory with the specified color.
// this won't update the display.
func (d *EInk) ResetFrameMemory(color byte) error {
	var (
		w = d.rect.Dx()
		h = d.rect.Dy()
	)

	if err := d.setMemoryArea(0, 0, w - 1, h - 1); err != nil {
		return err
	}
	if err := d.setMemoryPointer(0, 0); err != nil {
		return err
	}
	if err := d.SendCommandArgs(writeRAM); err != nil {
		return err
	}

	// send the color data
	for i := 0; i < (w / 8 * h); i++ {
		if err := d.SendData(color); err != nil {
			return err
		}
	}

	return nil
}

// Refresh updates the EInk display.
func (d *EInk) Refresh() error {
	if err := d.SendCommandArgs(displayUpdateControl2, 0xC4); err != nil {
		return err
	}
	
	if err := d.SendCommandArgs(masterActivation); err != nil {
		return err
	}

	if err := d.SendCommandArgs(terminateFrameReadWrite); err != nil {
		return err
	}

	d.waitUntilIdle()
	return nil
}

// Clear clears the EInk display.
func (d *EInk) Clear() error {
	return d.ResetFrameMemory(0xFF)
}

// ClearAndRefresh clears the EInk display and triggers update of the frame.
func (d *EInk) ClearAndRefresh() error {
	if err := d.Clear(); err != nil {
		return err
	}

	return d.Refresh()
}

// Sleep puts EInk display to deep-sleep mode to save power.
// Use Reset() to awaken and Init to re-initialize the device.
func (d *EInk) Sleep() error {
	if err := d.SendCommandArgs(deepSleepMode); err != nil {
		return err
	}

	d.waitUntilIdle()
	return nil
}

// Reset performs hardware reset of the EInk display.
func (d *EInk) Reset() (err error) {
	if err = d.rst.Out(gpio.High); err != nil {
		return
	}
	time.Sleep(200 * time.Millisecond)

	if err = d.rst.Out(gpio.Low); err != nil {
		return
	}
	time.Sleep(200 * time.Millisecond)

	if err = d.rst.Out(gpio.High); err != nil {
		return
	}
	time.Sleep(200 * time.Millisecond)

	return
}

// ColorModel implements display.Drawer.
// It is a one bit color model, as implemented by image1bit.Bit.
func (d *EInk) ColorModel() color.Model {
	return image1bit.BitModel
}

// Bounds implements display.Drawer. Min is guaranteed to be {0, 0}.
func (d *EInk) Bounds() image.Rectangle {
	return image.Rect(0, 0, d.rect.Dy(), d.rect.Dx())
}

// SendCommandArgs overrides periphery.SPI send command with args method
// by additionally sending signals to DC and CS GPIO pins.
func (d *EInk) SendCommandArgs(cmd byte, data ...byte) error {
	if !d.Active() {
		return nil
	}

	if err := d.SendCommand(cmd); err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	return d.SendData(data...)
}

// SendCommand overrides periphery.SPI send command method
// by additionally sending signals to DC and CS GPIO pins.
func (d *EInk) SendCommand(cmd byte) (err error) {
	if !d.Active() {
		return
	}

	if err := d.dc.Out(gpio.Low); err != nil {
		return errors.Wrapf(err, "error during sending %s signal to %s", d.dc, gpio.Low)
	}

	if err := d.cs.Out(gpio.Low); err != nil {
		return errors.Wrapf(err, "error during sending %s signal to %s", d.cs, gpio.Low)
	}

	defer func() {
		if err2 := d.cs.Out(gpio.High); err2 != nil {
			err = errors.Errorf("multiply errors during sending command to SPI device: %s; %s", err, err2)
		}
	}()

	return d.SPI.SendCommand(cmd)
}

// SendData overrides periphery.SPI send data method
// by additionally sending signals to DC and CS GPIO pins.
func (d *EInk) SendData(data ...byte) (err error) {
	if !d.Active() {
		return nil
	}

	if len(data) == 0 {
		return nil
	}

	if err := d.cs.Out(gpio.Low); err != nil {
		return errors.Wrapf(err, "error during sending %s signal to %s", d.cs, gpio.Low)
	}

	if err := d.dc.Out(gpio.High); err != nil {
		return errors.Wrapf(err, "error during sending %s signal to %s", d.dc, gpio.High)
	}

	defer func() {
		if err2 := d.cs.Out(gpio.High); err2 != nil {
			err = errors.Errorf("multiply errors during sending data to SPI device: %s; %s", err, err2)
		}
	}()

	return d.SPI.SendData(data...)
}

// init performs sequence of commands to initialise EInk display chip.
func (d *EInk) init() error {
	if err := d.Reset(); err != nil {
		return err
	}

	if err := d.SendCommandArgs(swReset); err != nil {
		return err
	}
	d.waitUntilIdle()

	if err := d.SendCommandArgs(autoWriteRamBW, 0xF7); err != nil {
		return err
	}
	d.waitUntilIdle()

	if err := d.SendCommandArgs(driverOutputControl,
		byte((d.config.Width - 1) & 0xFF),
		byte(((d.config.Height - 1) >> 8) & 0xFF), 0x01); err != nil {
		return err
	}

	if err := d.SendCommandArgs(boosterSoftStartControl, 0xAE, 0xC7, 0xC3, 0xC0, 0x40); err != nil {
		return err
	}

	if err := d.SendCommandArgs(dataEntryModeSetting, 0x01); err != nil {
		return err
	}

	if err := d.setMemoryArea(0, 0, d.config.Width - 1, d.config.Height - 1); err != nil {
		return err
	}

	if err := d.SendCommandArgs(borderWaveformControl, 0x01); err != nil {
		return err
	}

	if err := d.SendCommandArgs(temperatureSensorControl, 0x80); err != nil {
		return err
	}

	if err := d.SendCommandArgs(displayUpdateControl2, 0xB1); err != nil {
		return err
	}

	if err := d.SendCommandArgs(masterActivation); err != nil {
		return err
	}

	d.waitUntilIdle()

	if err := d.SendCommandArgs(displayUpdateControl2, 0xB1); err != nil {
		return err
	}

	if err := d.SendCommandArgs(masterActivation); err != nil {
		return err
	}

	return d.setMemoryPointer(0, 0)
}

func rotateBitMap(bitMap *image1bit.VerticalLSB) *image1bit.VerticalLSB {
	next := image1bit.NewVerticalLSB(image.Rect(
		bitMap.Rect.Min.Y,
		bitMap.Rect.Min.X,
		bitMap.Rect.Max.Y,
		bitMap.Rect.Max.X,
	))

	for x := bitMap.Rect.Min.X; x < bitMap.Rect.Max.X; x++ {
		for y := bitMap.Rect.Min.Y; y < bitMap.Rect.Max.Y; y++ {
			bit := bitMap.BitAt(x, y)
			next.SetBit(bitMap.Rect.Max.Y - 1 - y, x, bit)
		}
	}

	return next
}

func (d *EInk) setMemoryPointer(x, y int) error {
	if err := d.SendCommandArgs(setRAMXAddressCounter, byte((x >> 3) & 0xFF)); err != nil {
		return err
	}

	if err := d.SendCommandArgs(setRAMYAddressCounter, byte(y & 0xFF), byte((y >> 8) & 0xFF)); err != nil {
		return err
	}

	d.waitUntilIdle()

	return nil
}

func (d *EInk) waitUntilIdle() {
	for d.busy.Read() == gpio.High {
		time.Sleep(100 * time.Millisecond)
	}
}

func (d *EInk) setMemoryArea(xStart, yStart, xEnd, yEnd int) error {
	if err := d.SendCommandArgs(setRAMXAddressStartEndPosition,
		byte((xStart >> 3) & 0xFF),
		byte((xEnd >> 3) & 0xFF),
	); err != nil {
		return err
	}

	return d.SendCommandArgs(setRAMYAddressStartEndPosition,
		byte(yStart & 0xFF),
		byte((yStart >> 8) & 0xFF),
		byte(yEnd & 0xFF),
		byte((yEnd >> 8) & 0xFF),
	)
}
