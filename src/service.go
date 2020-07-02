package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/thingsplex/fronius/fronius"

	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/edgeapp"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/fronius/handler"
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
	system := fronius.System{}
	// state := fronius.State{}

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

	fimpRouter := handler.NewFromFimpRouter(mqtt, appLifecycle, configs, states)
	fimpRouter.Start()

	PollTime := configs.PollTimeSec
	for {
		appLifecycle.WaitForState("main", edgeapp.AppStateRunning)
		log.Info("--------------Starting ticker---------------")
		ticker := time.NewTicker(time.Duration(PollTime) * time.Second)
		for ; true; <-ticker.C {
			log.Debug(states.Connected)
			if states.Connected {
				req, err := http.NewRequest("GET", fronius.GetRealTimeDataURL(configs.Host), nil)
				if err != nil {
					log.Error(fmt.Errorf("Can't get measurements - ", err))
				}
				log.Debug("")
				log.Debug(req)
				log.Debug("")
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					log.Error(err)
				} else {
					measurements, err := system.NewResponse(resp)
					if err != nil {
						log.Error(err)
					}
					log.Debug(resp)
					log.Debug("")
					fimpRouter.SendMeasurements(system.Head.RequestArguments.DeviceId, measurements)
				}
			} else {
				log.Debug("-------NOT CONNECTED------")
				// Do nothing
			}
			states.SaveToFile()
		}
		appLifecycle.WaitForState(edgeapp.AppStateNotConfigured, "main")
	}

	mqtt.Stop()
	time.Sleep(5 * time.Second)
}
