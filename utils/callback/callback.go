package callback

import (
	"../log"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
	"../../config"
)

var netTransport = &http.Transport{
	Dial: (&net.Dialer{
		Timeout: 5 * time.Second,
	}).Dial,
	TLSHandshakeTimeout: 5 * time.Second,
	MaxIdleConnsPerHost: 5,
}

var httpCallbackClient = &http.Client{
	Timeout:   time.Second * 10,
	Transport: netTransport,
}

var (
	InStockConfirmCallbackUrl  = config.CALLBACK_URL + "/callback/devicein?actionID=%d&terminalID=%d&deviceID=%d&result=%d&slotID=%d"
	OUtStockConfirmCallbackUrl = config.CALLBACK_URL + "/callback/deviceout?actionID=%d&terminalID=%d&deviceID=%d&result=%d&slotID=%d"
	SyncTerminalCallbackUrl    = config.CALLBACK_URL + "/callback/initTerminal?terminalID=%d"
	HTTPNotOKError = errors.New("HTTP status not OK")
)

type QueryRsp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func PhraseHttpCallback(url, method string, query_dict map[string]interface{}, obj interface{}) error {
	post_data, err := json.Marshal(query_dict)
	if err != nil {
		log.Error("postdata marshal fail", err, query_dict)
		return errors.New("postdata marshal fail")
	}
	var resp *http.Response
	if method == "POST" {
		resp, err = httpCallbackClient.Post(url, "application/json", strings.NewReader(string(post_data)))
	} else {
		resp, err = httpCallbackClient.Get(url)
	}
	if err != nil {
		log.Error(url, " send request error, url|err:", url, err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Println(url, "http status error, url|http status:", url, resp.Status)
		return HTTPNotOKError
	}
	resp_data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(url, "read response data error, url|err:", url, err)
		return err
	}
	err = json.Unmarshal(resp_data, &obj)
	if err != nil {
		log.Error("res unmarshal err", err, string(resp_data))
		return errors.New("res unmarshal err")
	}
	return nil
}

func InStockCallBack(actionId, terminalId, deviceId, result, slotId uint32) {
	url := fmt.Sprintf(InStockConfirmCallbackUrl, actionId, terminalId, deviceId, result, slotId)
	qd := make(map[string]interface{})
	res := QueryRsp{}
	err := PhraseHttpCallback(url, "POST", qd, &res)
	if err != nil {
		log.Error("InStockCallBack err", actionId, terminalId, deviceId, result, err)
		return
	}
	log.Info("InStockCallBack res", url, res)
}

func OutStockCallBack(actionId, terminalId, deviceId, result, slotId uint32) {
	url := fmt.Sprintf(OUtStockConfirmCallbackUrl, actionId, terminalId, deviceId, result, slotId)
	qd := make(map[string]interface{})
	res := QueryRsp{}
	err := PhraseHttpCallback(url, "POST", qd, &res)
	if err != nil {
		log.Error("OutStockCallBack err", actionId, terminalId, deviceId, result, err)
		return
	}
	log.Info("OutStockCallBack res", url, res)
}


func SyncTerminal(terminalId uint32){
	url := fmt.Sprintf(SyncTerminalCallbackUrl, terminalId)
	qd := make(map[string]interface{})
	res := QueryRsp{}
	err := PhraseHttpCallback(url, "GET", qd, &res)
	if err != nil {
		log.Error("SyncTerminal err", terminalId, err)
		return
	}
	log.Info("SyncTerminal res", res)
}