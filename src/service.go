package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/edgeapp"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/fronius/model"
	"github.com/thingsplex/fronius/utils"
)

func main() {
	var workDir string
	flag.StringVar(&workDir, "c", "", "Work dir")
	flag.Parse()
	if workDir == "" {
		workDir = "./"
	} else {
		fmt.Println("Work dir ", workDir)
	}
	appLifecycle := edgeapp.NewAppLifecycle()
	configs := model.NewConfigs(workDir)
	states := model.NewStates(workDir)
	err := configs.LoadFromFile()
	if err != nil {
		appLifecycle.SetAppState(edgeapp.AppStateStartupError, nil)
		fmt.Print(err)
		panic("Can't load config file.")
	}

	err = states.LoadFromFile()
	if err != nil {
		appLifecycle.SetConfigState(edgeapp.ConfigStateNotConfigured)
		fmt.Print(err)
		panic("Not able to load state")
	}

	utils.SetupLog(configs.LogFile, configs.LogLevel, configs.LogFormat)
	log.Info("--------------Starting fronius----------------")
	appLifecycle.SetAppState(edgeapp.AppStateStarting, nil)
	appLifecycle.SetAuthState(edgeapp.AuthStateNotAuthenticated)
	appLifecycle.SetConfigState(edgeapp.ConfigStateNotConfigured)
	appLifecycle.SetConnectionState(edgeapp.ConnStateDisconnected)

	log.Info("Work directory : ", configs.WorkDir)

	mqtt := fimpgo.NewMqttTransport(configs.MqttServerURI, configs.MqttClientIdPrefix, configs.MqttUsername, configs.MqttPassword, true, 1, 1)
	err = mqtt.Start()

	if err != nil {
		log.Error("Can't connect to broker. Error: ", err.Error())
	} else {
		log.Info("----------------Connected------------------")
	}
	defer mqtt.Stop()
	appLifecycle.SetAppState(edgeapp.AppStateRunning, nil)

	if err := edgeapp.NewSystemCheck().WaitForInternet(5 * time.Minute); err == nil {
		log.Info("<main> Internet connection - OK")
	} else {
		log.Error("<main> Internet connection - ERROR")
	}
	if states.IsConfigured() && err == nil {
		appLifecycle.SetConfigState(edgeapp.ConfigStateConfigured)
		appLifecycle.SetConnectionState(edgeapp.ConnStateConnected)
		appLifecycle.SetAuthState(edgeapp.AuthStateAuthenticated)
	} else {
		appLifecycle.SetConfigState(edgeapp.ConfigStateNotConfigured)
		appLifecycle.SetAuthState(edgeapp.AuthStateNotAuthenticated)
	}

	fimpHandler := handler.NewFimpFroniusHandler(mqtt, configs.StateDir, appLifecycle)
	fimpHandler.Start(configs.PollTimeSec)
	log.Info("-------------------Starting handler-----------------")

	mqtt.Subscribe("pt:j1/mt:cmd/rt:ad/rn:fronius/ad:1")
	mqtt.Subscribe("pt:j1/mt:evt/rt:dev/rn:fronius/ad:1/#")
	log.Info("Subscribing to topic: pt:j1/mt:cmd/rt:ad/rn:fronius/ad:1")
	log.Info("Subscribing to topic: pt:j1/mt:evt/rt:dev/rn:fronius/ad:1/#")

	select {}

	mqtt.Stop()
}
