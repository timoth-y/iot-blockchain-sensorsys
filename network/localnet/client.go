package localnet

import (
	"context"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/timoth-y/chainmetric-iot/drivers/periphery"
	"github.com/timoth-y/chainmetric-iot/shared"
)

// Client defines the interface for low range network communication.
type Client struct {
	dev *periphery.Bluetooth
}

var (
	client *Client

	// Channels exposes available channel for local network communication.
	Channels = struct {
		Geo *GeoLocationChannel
	} {
		Geo: newGeoLocationChannel(),
	}
)

// Init performs initialisation of the Client.
func Init(name string) error {
	client = &Client{
		dev: periphery.NewBluetooth(),
	}

	if !viper.GetBool("bluetooth.enabled") {
		return errors.New("localnet unavailable since bluetooth does not enabled")
	}

	if err := client.dev.Init(
		periphery.WithDeviceName(name),
		periphery.WithAdvertisementServices(Channels.Geo.uuid),
	); err != nil {
		return errors.Wrap(err, "failed to prepare Bluetooth driver")
	}

	Channels.Geo.init()

	if err := Channels.Geo.expose(client.dev); err != nil {
		return errors.Wrap(err, "failed to expose client to geo channel")
	}

	return nil
}

// Pair performs pairing via Bluetooth.
func Pair(ctx context.Context) error {
	if !viper.GetBool("bluetooth.enabled") {
		return errors.New("advertising unavailable since bluetooth does not enabled")
	}

	shared.Logger.Debug("Bluetooth pairing started")

	if err := client.dev.Advertise(ctx); err != nil {
		return errors.Wrap(err, "failed to advertise device via Bluetooth")
	}

	return nil
}

// SetDeviceName sets new `name` for identifying device on local network.
func SetDeviceName(name string) {
	client.dev.ApplyOptions(periphery.WithDeviceName(name))
}

// Close closes local network connection and clears allocated resources.
func Close() error {
	client = nil
	return nil
}
