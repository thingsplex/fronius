package fronius

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	api          = "solar_api/v1"
	getInvRtData = "GetInverterRealtimeData.cgi"
	scope        = "Scope=System"
	batteries    = "config/batteries"
	exportLimit  = "config/exportlimit"
	readable     = "components/cache/readable"
	powerflow    = "status/powerflow"
)

type System struct {
	Head struct {
		RequestArguments struct {
			DataCollection string `json:"DataCollection"`
			DeviceClass    string `json:"DeviceClass"`
			DeviceId       string `json:"DeviceId"`
			Scope          string `json:"Scope"`
		} `json:"RequestArguments"`
		Status struct {
			Code        int32  `json:"Code"`
			Reason      string `json:"Reason"`
			UserMessage string `json:"UserMessage"`
		} `json:"Status"`
		Timestamp string `json:"Timestamp"`
	} `json:"Head"`
	Body struct {
		Data struct {
			EnergyDay struct {
				Unit  string `json:"Unit"`
				Value struct {
					Value float64 `json:"1"`
				} `json:"Values"`
			} `json:"DAY_ENERGY"`
			DeviceStatus struct {
				ErrorCode              int32
				LEDColor               int32
				LEDState               int32
				MgmtTimerRemainingTime int32
				StateToReset           bool
				StatusCode             int32
			} `json:"DeviceStatus"`
			Freq struct {
				Unit  string `json:"Unit"`
				Value struct {
					Value float64 `json:"1"`
				} `json:"Values"`
			} `json:"FAC"`
			CurrentAC struct {
				Unit  string `json:"Unit"`
				Value struct {
					Value float64 `json:"1"`
				} `json:"Values"`
			} `json:"IAC"`
			CurrentDC struct {
				Unit  string `json:"Unit"`
				Value struct {
					Value float64 `json:"1"`
				} `json:"Values"`
			} `json:"IDC"`
			Power struct {
				Unit  string `json:"Unit"`
				Value struct {
					Value float64 `json:"1"`
				} `json:"Values"`
			} `json:"PAC"`
			EnergyTotal struct {
				Unit  string `json:"Unit"`
				Value struct {
					Value float64 `json:"1"`
				} `json:"Values"`
			} `json:"TOTAL_ENERGY"`
			VoltageAC struct {
				Unit  string `json:"Unit"`
				Value struct {
					Value float64 `json:"1"`
				} `json:"Values"`
			} `json:"UAC"`
			VoltageDC struct {
				Unit  string `json:"Unit"`
				Value struct {
					Value float64 `json:"1"`
				} `json:"Values"`
			} `json:"UDC"`
			EnergyYear struct {
				Unit  string `json:"Unit"`
				Value struct {
					Value float64 `json:"1"`
				} `json:"Values"`
			} `json:"YEAR_ENERGY"`
		} `json:"Data"`
	} `json:"Body"`
}

type SystemHybrid struct {
	Body struct {
		Data struct {
			Pac struct {
				Unit  string `json:"Unit"`
				Value struct {
					Num1 float64 `json:"1"`
				} `json:"Value"`
			} `json:"PAC"`
		} `json:"Data"`
	} `json:"Body"`
	Head struct {
		RequestArguments struct {
			DeviceClass string `json:"DeviceClass"`
			Scope       string `json:"Scope"`
		} `json:"RequestArguments"`
		Status struct {
			Code        int    `json:"Code"`
			Reason      string `json:"Reason"`
			UserMessage string `json:"UserMessage"`
		} `json:"Status"`
		Timestamp string `json:"Timestamp"`
	} `json:"Head"`
}

type Powerflow struct {
	Common struct {
		Datestamp string `json:"datestamp"`
		Timestamp string `json:"timestamp"`
	} `json:"common"`
	Inverters []struct {
		BatMode float64 `json:"BatMode"`
		Cid     int     `json:"CID"`
		Dt      int     `json:"DT"`
		ID      int     `json:"ID"`
		P       float64 `json:"P"`
		Soc     float64 `json:"SOC"`
	} `json:"inverters"`
	Site struct {
		BackupMode         bool        `json:"BackupMode"`
		BatteryStandby     bool        `json:"BatteryStandby"`
		EDay               interface{} `json:"E_Day"`
		ETotal             interface{} `json:"E_Total"`
		EYear              interface{} `json:"E_Year"`
		MLoc               int         `json:"MLoc"`
		Mode               string      `json:"Mode"`
		PAkku              float64     `json:"P_Akku"`
		PGrid              float64     `json:"P_Grid"`
		PLoad              float64     `json:"P_Load"`
		PPv                float64     `json:"P_PV"`
		RelAutonomy        float64     `json:"rel_Autonomy"`
		RelSelfConsumption float64     `json:"rel_SelfConsumption"`
	} `json:"site"`
	Version string `json:"version"`
}

type State struct {
	Value float64
	Unit  string
}

func GetRealTimeDataURL(host string) string {
	url := fmt.Sprintf("%s%s%s%s%s%s%s", host, "/", api, "/", getInvRtData, "?", scope)
	return url
}

// func

func (sys System) NewResponse(httpresp *http.Response) (system System, err error) {
	body, err := ioutil.ReadAll(httpresp.Body)
	if err != nil {
		// handle err
		return system, err
	}

	err = json.Unmarshal(body, &system)
	return system, err
}

func (sysh SystemHybrid) NewHybridResponse(httpresp *http.Response) (systemh SystemHybrid, err error) {
	body, err := ioutil.ReadAll(httpresp.Body)
	if err != nil {
		// handle err
		return systemh, err
	}

	err = json.Unmarshal(body, &systemh)
	return systemh, err
}

func (pow Powerflow) NewPowerflowResponse(httpresp *http.Response) (powerf Powerflow, err error) {
	body, err := ioutil.ReadAll(httpresp.Body)
	if err != nil {
		// handle err
		return powerf, err
	}

	err = json.Unmarshal(body, &powerf)
	return powerf, err
}

func (st State) CurrentPowerHybrid(powf Powerflow) State {
	for _, inv := range powf.Inverters {
		st.Value += inv.P
	}
	st.Unit = "W"
	return st
}

func (st State) CurrentPower(sys System) State {
	st.Value = sys.Body.Data.Power.Value.Value
	st.Unit = sys.Body.Data.Power.Unit
	return st
}

func (st State) EnergyDay(sys System) State {
	st.Value = sys.Body.Data.EnergyDay.Value.Value
	st.Unit = sys.Body.Data.EnergyDay.Unit
	return st
}

func (st State) EnergyYear(sys System) State {
	st.Value = sys.Body.Data.EnergyYear.Value.Value
	st.Unit = sys.Body.Data.EnergyYear.Unit
	return st
}

func (st State) EnergyTotal(sys System) State {
	st.Value = sys.Body.Data.EnergyTotal.Value.Value
	st.Unit = sys.Body.Data.EnergyTotal.Unit
	return st
}

func (st State) GetSystems() (bool, error) {
	// value := url.Values{
	// 	"fields": []string{""}
	// }
	return true, nil
}

func (st State) Resolve(ip string, doneChannel chan bool) {
	addresses, err := net.LookupAddr(ip)
	if err != nil {
		log.Debug("resolve error")
		log.Debug(err)
	}
	if err == nil {
		log.Debug("in the middle")
		log.Debug(fmt.Printf("%s - %s", ip, strings.Join(addresses, ", ")))
	}
	doneChannel <- true
}
