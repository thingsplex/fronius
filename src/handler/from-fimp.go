package handler

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

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
	fc := FromFimpRouter{inboundMsgCh: make(fimpgo.MessageCh, 5), mqt: mqt, appLifecycle: appLifecycle, configs: configs, state: states}
	fc.mqt.RegisterChannel("ch1", fc.inboundMsgCh)
	return &fc
}

func (fc *FromFimpRouter) Start() {

	fc.mqt.Subscribe(fmt.Sprintf("pt:j1/+/rt:dev/rn:%s/ad:1/#", model.ServiceName))
	fc.mqt.Subscribe(fmt.Sprintf("pt:j1/+/rt:ad/rn:%s/ad:1", model.ServiceName))
	log.Debug(fmt.Sprintf("Subscribing to topic: pt:j1/+/rt:dev/rn:%s/ad:1/#", model.ServiceName))
	log.Debug(fmt.Sprintf("Subscribing to topic: pt:j1/+/rt:ad/rn:%s/ad:1", model.ServiceName))

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
		fc.configs.Host = conf.Host
		fc.configs.Type = conf.Type
		fc.configs.Value1 = conf.Value1
		fc.configs.Value2 = conf.Value2
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

	case "cmd.system.forced_battery_storage_prestart":
		url := "http://" + fc.configs.Host
		digestPost(url, "/config/batteries", []byte("{\"HYB_EVU_CHARGEFROMGRID\":true}"))
		digestPost(url, "/config/batteries", []byte("{\"HYB_EM_POWER\":-50000,\"HYB_EM_MODE\":1}"))
		val := model.ButtonActionResponse{
			Operation:       "cmd.system.forced_battery_storage_prestart",
			OperationStatus: "ok",
			Next:            "reload",
			ErrorCode:       "",
			ErrorText:       "",
		}
		msg := fimpgo.NewMessage("evt.app.config_action_report", model.ServiceName, fimpgo.VTypeObject, val, nil, nil, newMsg.Payload)
		if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
			fc.mqt.Publish(adr, msg)
		}

	case "cmd.system.forced_battery_storage":
		url := "http://" + fc.configs.Host
		digestPost(url, "/config/batteries", []byte("{\"HYB_EM_POWER\":50000,\"HYB_EM_MODE\":1}"))
		val := model.ButtonActionResponse{
			Operation:       "cmd.system.forced_battery_storage",
			OperationStatus: "ok",
			Next:            "reload",
			ErrorCode:       "",
			ErrorText:       "",
		}
		msg := fimpgo.NewMessage("evt.app.config_action_report", model.ServiceName, fimpgo.VTypeObject, val, nil, nil, newMsg.Payload)
		if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
			fc.mqt.Publish(adr, msg)
		}

	case "cmd.system.forced_battery_storage_finished":
		url := "http://" + fc.configs.Host
		digestPost(url, "/config/batteries", []byte("{\"HYB_EVU_CHARGEFROMGRID\":false}"))
		digestPost(url, "/config/batteries", []byte("{\"HYB_EM_POWER\":0,\"HYB_EM_MODE\":1}"))
		val := model.ButtonActionResponse{
			Operation:       "cmd.system.forced_battery_storage_finished",
			OperationStatus: "ok",
			Next:            "reload",
			ErrorCode:       "",
			ErrorText:       "",
		}
		msg := fimpgo.NewMessage("evt.app.config_action_report", model.ServiceName, fimpgo.VTypeObject, val, nil, nil, newMsg.Payload)
		if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
			fc.mqt.Publish(adr, msg)
		}

	case "cmd.system.excess_solar_production_disabled":
		url := "http://" + fc.configs.Host
		digestPost(url, "/config/exportlimit", []byte("{\"DPL_ON\":true,\"DPL_WPEAK\":5000,\"DPL_WLIM_USE_ABS\":true,\"DPL_WLIM_ABS\":0}"))
		val := model.ButtonActionResponse{
			Operation:       "cmd.system.excess_solar_production_disabled",
			OperationStatus: "ok",
			Next:            "reload",
			ErrorCode:       "",
			ErrorText:       "",
		}
		msg := fimpgo.NewMessage("evt.app.config_action_report", model.ServiceName, fimpgo.VTypeObject, val, nil, nil, newMsg.Payload)
		if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
			fc.mqt.Publish(adr, msg)
		}

	case "cmd.system.excess_solar_production_enabled":
		url := "http://" + fc.configs.Host
		digestPost(url, "/config/exportlimit", []byte("{\"DPL_ON\":false}"))
		val := model.ButtonActionResponse{
			Operation:       "cmd.system.excess_solar_production_enabled",
			OperationStatus: "ok",
			Next:            "reload",
			ErrorCode:       "",
			ErrorText:       "",
		}
		msg := fimpgo.NewMessage("evt.app.config_action_report", model.ServiceName, fimpgo.VTypeObject, val, nil, nil, newMsg.Payload)
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
		inclReport := model.SendInclusionReport()

		msg := fimpgo.NewMessage("evt.thing.inclusion_report", "fronius", fimpgo.VTypeObject, inclReport, nil, nil, nil)
		adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: "fronius", ResourceAddress: "1"}
		fc.mqt.Publish(&adr, msg)

	case "cmd.thing.inclusion":
		inclReport := model.SendInclusionReport()

		msg := fimpgo.NewMessage("evt.thing.inclusion_report", "fronius", fimpgo.VTypeObject, inclReport, nil, nil, nil)
		adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: "fronius", ResourceAddress: "1"}
		fc.mqt.Publish(&adr, msg)

	case "cmd.app.uninstall":
		val := map[string]interface{}{
			"address": "1",
		}
		msg := fimpgo.NewMessage("evt.thing.exclusion_report", "fronius", fimpgo.VTypeObject, val, nil, nil, newMsg.Payload)
		msg.Source = "fronius"
		adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: "fronius", ResourceAddress: "1"}
		fc.mqt.Publish(&adr, msg)

	case "cmd.thing.delete":
		// remove device from network
		val, err := newMsg.Payload.GetStrMapValue()
		if err != nil {
			log.Error("Wrong msg format")
			return
		}
		fc.configs.Host = "host_ip"
		fc.appLifecycle.SetConfigState(edgeapp.ConfigStateNotConfigured)
		deviceId, ok := val["address"]
		if ok {
			log.Info("Deleting device")
			exclReport := make(map[string]string)
			exclReport["address"] = deviceId

			msg := fimpgo.NewMessage("evt.thing.exclusion_report", "fronius", fimpgo.VTypeObject, exclReport, nil, nil, nil)
			adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: "fronius", ResourceAddress: "1"}
			fc.mqt.Publish(&adr, msg)
		} else {
			log.Error("Incorrect address")

		}
	}
}

func digestPost(host string, uri string, postBody []byte) bool {
	url := host + uri
	log.Debug("url: ", url)
	method := "POST"
	req, err := http.NewRequest(method, url, nil)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		log.Debug("Recieved status code '%v' auth skipped", resp.StatusCode)
		return true
	}
	log.Debug("digestparts")
	digestParts := digestParts(resp)
	log.Debug("digestparts finished")
	digestParts["uri"] = uri
	digestParts["method"] = method
	digestParts["username"] = "technician"
	digestParts["password"] = "Solcelle2021"

	log.Debug("digestparts: ", digestParts)
	req, err = http.NewRequest(method, url, bytes.NewBuffer(postBody))
	req.Header.Set("Authorization", getDigestAuthrization(digestParts))
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		log.Debug("error1: ", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Debug("error2: ", err)
		}
		log.Debug("response body: ", string(body))
		return false
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debug("error2: ", err)
	}
	log.Debug("response body: ", string(body))
	return true
}

func digestParts(resp *http.Response) map[string]string {
	result := map[string]string{}
	if len(resp.Header["X-Www-Authenticate"]) > 0 {
		wantedHeaders := []string{"nonce", "realm", "qop"}
		responseHeaders := strings.Split(resp.Header["X-Www-Authenticate"][0], ",")
		for _, r := range responseHeaders {
			for _, w := range wantedHeaders {
				if strings.Contains(r, w) {
					result[w] = strings.Split(r, `"`)[1]
				}
			}
		}
		// case "cmd.thing.delete":
		// 	// remove device from network
		// 	val, err := newMsg.Payload.GetStrMapValue()
		// 	if err != nil {
		// 		log.Error("Wrong msg format")
		// 		return
		// 	}
		// 	deviceId, ok := val["address"]
		// 	if ok {
		// 		log.Info(deviceId)
		// 		exclReport := make(map[string]string)
		// 		exclReport["address"] = deviceId

		// 		msg := fimpgo.NewMessage("evt.thing.exclusion_report", "fronius", fimpgo.VTypeObject, exclReport, nil, nil, nil)
		// 		adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: "fronius", ResourceAddress: "1"}
		// 		fc.mqt.Publish(&adr, msg)
		// 	} else {
		// 		log.Error("Incorrect address")

		// 	}
	}
	return result
}

func getMD5(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func getCnonce() string {
	b := make([]byte, 8)
	io.ReadFull(rand.Reader, b)
	return fmt.Sprintf("%x", b)[:16]
}

func getDigestAuthrization(digestParts map[string]string) string {
	d := digestParts
	ha1 := getMD5(d["username"] + ":" + d["realm"] + ":" + d["password"])
	ha2 := getMD5(d["method"] + ":" + d["uri"])
	nonceCount := 00000001
	cnonce := getCnonce()
	response := getMD5(fmt.Sprintf("%s:%s:%v:%s:%s:%s", ha1, d["nonce"], nonceCount, cnonce, d["qop"], ha2))
	authorization := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", cnonce="%s", nc="%v", qop="%s", response="%s"`,
		d["username"], d["realm"], d["nonce"], d["uri"], cnonce, nonceCount, d["qop"], response)
	log.Debug("authorization: ", authorization)
	return authorization
}
