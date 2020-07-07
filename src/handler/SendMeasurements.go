package handler

import (
	"github.com/futurehomeno/fimpgo"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/fronius/fronius-api"
)

func (fc *FromFimpRouter) SendMeasurements(addr string, meas fronius.System) {
	state := fronius.State{}
	val := make(map[string]float64)
	val["p_export"] = state.CurrentPower(meas).Value
	val["last_e_export"] = state.EnergyDay(meas).Value / 1000

	msg := fimpgo.NewMessage("evt.meter_ext.report", "meter_elec", "float_map", val, nil, nil, nil)
	msg.Source = "fronius"
	adr, _ := fimpgo.NewAddressFromString("pt:j1/mt:evt/rt:dev/rn:fronius/ad:1/sv:meter_elec/ad:1")
	fc.mqt.Publish(adr, msg)
	log.Debug("Energy message sent")
}
