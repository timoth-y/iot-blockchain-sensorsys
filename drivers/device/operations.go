package device

import (
	"time"

	"github.com/pkg/errors"
	"github.com/timoth-y/iot-blockchain-contracts/models"

	"github.com/timoth-y/iot-blockchain-sensorsys/model"
	"github.com/timoth-y/iot-blockchain-sensorsys/shared"
)

func (d *Device) Operate() {
	d.reader.RegisterSensors(d.SupportedSensors()...)

	for _, request := range d.requests.Get() {
		d.actOnRequest(request)
	}

	d.reader.Process()
}

func (d *Device) actOnRequest(request *readingsRequest) {
	var (
		handler = func(readings model.MetricReadings) {
			d.postReadings(request.assetID, readings)
		}
	)

	if request.period.Seconds() == 0 {
		d.reader.SendRequest(handler, request.metrics...)
		return
	}

	request.cancel = d.reader.SubscribeReceiver(handler, request.period, request.metrics...)
}

func (d *Device) postReadings(assetID string, readings model.MetricReadings) {
	var (
		contract = d.client.Contracts.Readings
	)

	if err := contract.Post(models.MetricReadings{
		AssetID: assetID,
		DeviceID: d.model.ID,
		Timestamp: time.Now(),
		Values: readings,
	}); err != nil {
		shared.Logger.Error(errors.Wrap(err, "failed to post readings"))
	}
}