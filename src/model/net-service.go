package model

import (
	"fmt"

	"github.com/futurehomeno/fimpgo/fimptype"
	"github.com/thingsplex/fronius/fronius"
)

// SendInclusionReport sends inclusion report for one system
func SendInclusionReport(nodeID int, SystemCollection []fronius.System) fimptype.ThingInclusionReport {
	var name, manufacturer string
	var deviceAddr string
	services := []fimptype.Service{}

	meterElecInterfaces := []fimptype.Interface{{
		Type:      "in",
		MsgType:   "cmd.meter.get_report",
		ValueType: "string",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.meter.reset",
		ValueType: "null",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.meter.report",
		ValueType: "float",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.meter_ext.report",
		ValueType: "float_map",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.meter_ext.get_report",
		ValueType: "null",
		Version:   "1",
	}}

	meterElecService := fimptype.Service{
		Name:    "meter_elec",
		Alias:   "meter_elec",
		Address: "/rt:dev/rn:fronius/ad:1/sv:meter_elec/ad:",
		Enabled: true,
		Groups:  []string{"ch_0"},
		Props: map[string]interface{}{
			"sup_units":         []string{"W", "kWh", "A", "V"},
			"sup_extended_vals": []string{"kWh", "W", "VAR", "Hz", "V", "A"},
		},
		Interfaces: meterElecInterfaces,
	}

	system := SystemCollection[nodeID]
	systemID := system.Head.RequestArguments.DeviceId

	manufacturer = "fronius"
	name = ""
	serviceAddress := fmt.Sprintf("%s", systemID)
	meterElecService.Address = meterElecService.Address + serviceAddress
	services = append(services, meterElecService)
	deviceAddr = fmt.Sprintf("%s", systemID)
	powerSource := "AC"

	inclReport := fimptype.ThingInclusionReport{
		IntegrationId:     "",
		Address:           deviceAddr,
		Type:              "",
		ProductHash:       manufacturer,
		CommTechnology:    "wifi",
		ProductName:       name,
		ManufacturerId:    manufacturer,
		DeviceId:          systemID,
		HwVersion:         "1",
		SwVersion:         "1",
		PowerSource:       powerSource,
		WakeUpInterval:    "-1",
		Security:          "",
		Tags:              nil,
		Groups:            []string{"ch_0"},
		PropSets:          nil,
		TechSpecificProps: nil,
		Services:          services,
	}

	return inclReport
}
