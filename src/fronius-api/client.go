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

type State struct {
	Value float64
	Unit  string
}

func GetRealTimeDataURL(host string) string {
	url := fmt.Sprintf("%s%s%s%s%s%s%s", host, "/", api, "/", getInvRtData, "?", scope)
	return url
}

func (sys System) NewResponse(httpresp *http.Response) (system System, err error) {
	body, err := ioutil.ReadAll(httpresp.Body)
	if err != nil {
		// handle err
		return system, err
	}

	err = json.Unmarshal(body, &system)
	return system, err
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
