package handler

import (
	"github.com/futurehomeno/fimpgo"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/fronius/fronius"
)

func (fc *FimpFroniusHandler) SendMeasurements(addr string, meas fronius.System) {
	state := fronius.State{}
	val := make(map[string]fronius.State)
	val["last_e_export"] = state.EnergyDay(meas)
	msg := fimpgo.NewMessage("evt.meter_ext.report", "meter_elec", "float_map", val, nil, nil, nil)
	msg.Source = "fronius"
	adr, _ := fimpgo.NewAddressFromString("pt:j1/mt:evt/rt:dev/rn:fronius/ad:1/sv:sensor_temp/ad:" + addr)
	fc.mqt.Publish(adr, msg)
	log.Debug("Energy message sent")
}
