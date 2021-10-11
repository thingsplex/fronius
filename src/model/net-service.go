package model

import (
	"fmt"

	"github.com/futurehomeno/fimpgo/fimptype"
)

// SendInclusionReport sends inclusion report for one system
func SendInclusionReport() fimptype.ThingInclusionReport {
	var name, manufacturer string
	var deviceAddr string
	services := []fimptype.Service{}

	inverterInterfaces := []fimptype.Interface{{
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

	inverterService := fimptype.Service{
		Name:    "inverter",
		Alias:   "inverter",
		Address: "/rt:dev/rn:fronius/ad:1/sv:inverter/ad:",
		Enabled: true,
		Groups:  []string{"ch_0"},
		Props: map[string]interface{}{
			"sup_units":         []string{"W", "kWh", "A", "V"},
			"sup_extended_vals": []string{"e_export", "last_e_export", "p_export", "freq", "u1", "u2", "u3", "i1", "i2", "i3"},
		},
		Interfaces: inverterInterfaces,
	}

	systemID := "1"

	manufacturer = "fronius"
	name = ""
	serviceAddress := fmt.Sprintf("%s", systemID)
	inverterService.Address = inverterService.Address + serviceAddress
	services = append(services, inverterService)
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

func SendHybridInclusionReport() fimptype.ThingInclusionReport {
	var name, manufacturer string
	var deviceAddr string
	services := []fimptype.Service{}

	inverterInterfaces := []fimptype.Interface{{
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

	batteryChargeInterfaces := []fimptype.Interface{{
		Type:      "out",
		MsgType:   "evt.meter_ext.report",
		ValueType: "float_map",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.meter_ext.get_report",
		ValueType: "null",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.mode.get_report",
		ValueType: "null",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.mode.report",
		ValueType: "string",
		Version:   "1",
	}}

	batteryInterfaces := []fimptype.Interface{{
		Type:      "in",
		MsgType:   "cmd.lvl.get_report",
		ValueType: "null",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.lvl.report",
		ValueType: "int",
		Version:   "1",
	}}

	inverterGridService := fimptype.Service{
		Name:    "inverter_grid_conn",
		Alias:   "inverter_grid_conn",
		Address: "/rt:dev/rn:fronius/ad:1/sv:inverter_grid_conn/ad:",
		Enabled: true,
		Groups:  []string{"ch_0"},
		Props: map[string]interface{}{
			"sup_units":         []string{"W"},                                            // ???
			"sup_extended_vals": []string{"p_export", "p_import", "e_export", "e_import"}, // ???
		},
		Interfaces: inverterInterfaces,
	}

	inverterSolarService := fimptype.Service{
		Name:    "inverter_solar_conn",
		Alias:   "inverter_solar_conn",
		Address: "/rt:dev/rn:fronius/ad:1/sv:inverter_solar_conn/ad:",
		Enabled: true,
		Groups:  []string{"ch_0"},
		Props: map[string]interface{}{
			"sup_units":         []string{"W"},        // ???
			"sup_extended_vals": []string{"p_export"}, // ???
		},
		Interfaces: inverterInterfaces,
	}

	batteryChargeService := fimptype.Service{
		Name:    "battery_charge_ctrl",
		Alias:   "battery_charge_ctrl",
		Address: "/rt:dev/rn:fronius/ad:1/sv:battery_charge_ctrl/ad:",
		Enabled: true,
		Groups:  []string{"ch_0"},
		Props: map[string]interface{}{
			"sup_modes": []string{"idle", "charging", "discharging"},
		},
		Interfaces: batteryChargeInterfaces,
	}

	batteryService := fimptype.Service{
		Name:       "battery",
		Alias:      "battery",
		Address:    "/rt:dev/rn:fronius/ad:1/sv:battery/ad:",
		Enabled:    true,
		Groups:     []string{"ch_0"},
		Props:      map[string]interface{}{},
		Interfaces: batteryInterfaces,
	}

	systemID := "1"

	manufacturer = "fronius"
	name = ""
	serviceAddress := fmt.Sprintf("%s", systemID)
	inverterGridService.Address = inverterGridService.Address + serviceAddress
	inverterSolarService.Address = inverterSolarService.Address + serviceAddress
	batteryChargeService.Address = batteryChargeService.Address + serviceAddress
	batteryService.Address = batteryService.Address + serviceAddress
	services = append(services, inverterGridService, inverterSolarService, batteryChargeService, batteryService)
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
