package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/timoth-y/chainmetric-sensorsys/drivers/device"
	"github.com/timoth-y/chainmetric-sensorsys/drivers/periphery"
	"github.com/timoth-y/chainmetric-sensorsys/drivers/sensors"
	"github.com/timoth-y/chainmetric-sensorsys/engine"
	"github.com/timoth-y/chainmetric-sensorsys/gateway/blockchain"
	"github.com/timoth-y/chainmetric-sensorsys/model/config"
	"github.com/timoth-y/chainmetric-sensorsys/shared"
)

var (
	Client = blockchain.NewBlockchainClient()
	// Display = display.NewEInk()
	Reader = engine.NewSensorsReader()
	Context = engine.NewContext(context.Background()).
		SetLogger(shared.Logger)
	Device = device.NewDevice().
		SetClient(Client).
		// SetDisplay(Display).
		SetReader(Reader)
)

func init() {
	shared.InitCore()
	periphery.Init()
}

func main() {
	done := make(chan struct{}, 1)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go run()
	go shutdown(quit, done)

	<-done
	shared.Logger.Info("Shutdown")
}

func run() {
	var (
		bc config.BlockchainConfig
		dc config.DisplayConfig
	)

	if err := shared.UnmarshalFromConfig("gateway", &bc); err != nil {
		shared.Logger.Fatal(errors.Wrap(err, "failed parse blockchain config"))
	}

	if err := shared.UnmarshalFromConfig("display", &dc); err != nil {
		shared.Logger.Fatal(errors.Wrap(err, "failed parse display config"))
	}

	if err := Client.Init(bc); err != nil {
		shared.Logger.Fatal(errors.Wrap(err, "failed initializing blockchain client"))
	}

	// if dc.Enabled {
	// 	if err := Display.Init(dc); err != nil {
	// 		shared.Logger.Fatal(errors.Wrap(err, "failed initializing display"))
	// 	}
	// }

	if err := Reader.Init(Context); err != nil {
		shared.Logger.Fatal(errors.Wrap(err, "failed initializing reader engine"))
	}

	if viper.GetBool("mocks.debug_env") {
		Device.RegisterStaticSensors(sensors.NewStaticSensorMock())
	}

	if err := Device.Init(); err != nil {
		shared.Logger.Fatal(errors.Wrap(err, "failed to initialize device"))
	}

	if err := Device.CacheBlockchainState(); err != nil {
		shared.Logger.Fatal(errors.Wrap(err, "failed to cache the state of blockchain"))
	}

	Device.WatchForBlockchainEvents()

	Device.Operate()
}

func shutdown(quit chan os.Signal, done chan struct{}) {
	<-quit
	shared.Logger.Info("Shutting down...")

	if Device != nil {
		if err := Device.Off(); err != nil {
			shared.Logger.Error(err)
		}

		if err := Device.Close(); err != nil {
			shared.Logger.Error(err)
		}
	}

	Client.Close()

	shared.CloseCore()

	close(done)
}
