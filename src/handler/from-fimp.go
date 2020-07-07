package handler

import (
	"fmt"
	"path/filepath"

	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/edgeapp"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/fronius/model"
)

type FromFimpRouter struct {
	inboundMsgCh fimpgo.MessageCh
	mqt          *fimpgo.MqttTransport
	instanceID   string
	appLifecycle *edgeapp.Lifecycle
	configs      *model.Configs
	state        *model.State
}

type ListReportRecord struct {
	Address        string `json:"address"`
	Alias          string `json:"alias"`
	WakeupInterval string `json:"wakeup_int"`
	PowerSource    string `json:"power_source"`
}

func NewFromFimpRouter(mqt *fimpgo.MqttTransport, appLifecycle *edgeapp.Lifecycle, configs *model.Configs, states *model.State) *FromFimpRouter {
	fc := FromFimpRouter{inboundMsgCh: make(fimpgo.MessageCh, 5), mqt: mqt, configs: configs, state: states}
	fc.mqt.RegisterChannel("ch1", fc.inboundMsgCh)
	return &fc
}

func (fc *FromFimpRouter) Start() {

	fc.mqt.Subscribe(fmt.Sprintf("pt:j1/+/rt:dev/rn:%s/ad:1/#", model.ServiceName))
	fc.mqt.Subscribe(fmt.Sprintf("pt:j1/+/rt:ad/rn:%s/ad:1", model.ServiceName))
	log.Debug("Subscribing to topic: pt:j1/+/rt:dev/rn:%s/ad:1/#")
	log.Debug("Subscribing to topic: pt:j1/+/rt:ad/rn:%s/ad:1")

	go func(msgChan fimpgo.MessageCh) {
		for {
			select {
			case newMsg := <-msgChan:
				fc.routeFimpMessage(newMsg)
			}
		}
	}(fc.inboundMsgCh)
}

func (fc *FromFimpRouter) routeFimpMessage(newMsg *fimpgo.Message) {

	log.Debug("New fimp msg")

	adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: model.ServiceName, ResourceAddress: "1"}

	switch newMsg.Payload.Type {

	case "cmd.app.get_manifest":
		mode, err := newMsg.Payload.GetStringValue()
		if err != nil {
			log.Error("Incorrect request format ")
			return
		}
		manifest := model.NewManifest()
		err = manifest.LoadFromFile(filepath.Join(fc.configs.GetDefaultDir(), "app-manifest.json"))
		if err != nil {
			log.Error("Failed to load manifest file .Error :", err.Error())
			return
		}
		if mode == "manifest_state" {
			manifest.AppState = fc.appLifecycle.GetAllStates()
			manifest.ConfigState = fc.configs
		}
		msg := fimpgo.NewMessage("evt.app.manifest_report", model.ServiceName, fimpgo.VTypeObject, manifest, nil, nil, newMsg.Payload)
		if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
			// if response topic is not set , sending back to default application event topic
			fc.mqt.Publish(adr, msg)
		}

	case "cmd.app.get_state":
		msg := fimpgo.NewMessage("evt.app.manifest_report", model.ServiceName, fimpgo.VTypeObject, fc.appLifecycle.GetAllStates(), nil, nil, newMsg.Payload)
		if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
			// if response topic is not set , sending back to default application event topic
			fc.mqt.Publish(adr, msg)
		}

	case "cmd.config.get_extended_report":

		msg := fimpgo.NewMessage("evt.config.extended_report", model.ServiceName, fimpgo.VTypeObject, fc.configs, nil, nil, newMsg.Payload)
		if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
			fc.mqt.Publish(adr, msg)
		}

	case "cmd.config.extended_set":
		conf := model.Configs{}
		err := newMsg.Payload.GetObjectValue(&conf)
		if err != nil {
			// TODO: This is an example . Add your logic here or remove
			log.Error("Can't parse configuration object")
			return
		}
		fc.configs.Param1 = conf.Param1
		fc.configs.Param2 = conf.Param2
		fc.configs.SaveToFile()
		log.Debugf("App reconfigured . New parameters : %v", fc.configs)
		// TODO: This is an example . Add your logic here or remove
		configReport := model.ConfigReport{
			OpStatus: "ok",
			AppState: fc.appLifecycle.GetAllStates(),
		}
		msg := fimpgo.NewMessage("evt.app.config_report", model.ServiceName, fimpgo.VTypeObject, configReport, nil, nil, newMsg.Payload)
		if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
			fc.mqt.Publish(adr, msg)
		}

	case "cmd.log.set_level":
		// Configure log level
		level, err := newMsg.Payload.GetStringValue()
		if err != nil {
			return
		}
		logLevel, err := log.ParseLevel(level)
		if err == nil {
			log.SetLevel(logLevel)
			fc.configs.LogLevel = level
			fc.configs.SaveToFile()
		}
		log.Info("Log level updated to = ", logLevel)

	case "cmd.network.get_all_nodes":
		// TODO: This is an example . Add your logic here or remove
	case "cmd.thing.get_inclusion_report":
		//nodeId , _ := newMsg.Payload.GetStringValue()
		// TODO: This is an example . Add your logic here or remove
	case "cmd.thing.inclusion":
		inclReport := model.SendInclusionReport(1, fc.state.Systems)

		msg := fimpgo.NewMessage("evt.thing.inclusion_report", "fronius", fimpgo.VTypeObject, inclReport, nil, nil, nil)
		adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: "fronius", ResourceAddress: "1"}
		fc.mqt.Publish(&adr, msg)

	case "cmd.thing.delete":
		// remove device from network
		val, err := newMsg.Payload.GetStrMapValue()
		if err != nil {
			log.Error("Wrong msg format")
			return
		}
		deviceId, ok := val["address"]
		if ok {
			// TODO: This is an example . Add your logic here or remove
			log.Info(deviceId)
		} else {
			log.Error("Incorrect address")

		}
	}
}
