package main

import (
	"net/http"
	"encoding/json"
	"io"
	"strconv"
	"time"
	"./utils/message"
	//"./utils/sqlutils"
)

func test_handler(w http.ResponseWriter, r *http.Request, args []string, pd []byte) {
	post_data := pd
	log.Info("post data", string(post_data))

	data := map[string]interface{}{
		"err": 0,

	}
	res, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	io.WriteString(w, string(res))
}

type QueryParams struct {
	SlotId byte
	DeviceId uint32
	ActionId uint32
	TerminalId uint32
}

func checkParams(r *http.Request) (*QueryParams, int64){
	_, exist := r.Form["slot_id"]
	if !exist {
		log.Println("slot_id not in request")
		return nil, -5
	}
	slot_id, err := strconv.ParseInt(string(r.Form["slot_id"][0]), 10, 32)
	if err != nil {
		log.Println("slot id convert int error", err, r.Form["slot_id"][0])
		return nil, -6
	}

	_, exist = r.Form["device_id"]
	if !exist {
		log.Println("device_id not in request")
		return nil, -7
	}
	device_id, err := strconv.ParseInt(string(r.Form["device_id"][0]), 10, 32)
	if err != nil {
		log.Println("device id convert int error", err, r.Form["device_id"][0])
		return nil, -8
	}

	_, exist = r.Form["action_id"]
	if !exist {
		log.Println("action_id not in request")
		return nil, -9
	}
	action_id, err := strconv.ParseInt(string(r.Form["action_id"][0]), 10, 32)
	if err != nil {
		log.Println("action id convert int error", err, r.Form["action_id"][0])
		return nil, -10
	}

	_, exist = r.Form["terminal_id"]
	if !exist {
		log.Println("terminal_id not in request")
		return nil, -11
	}
	terminal_id, err := strconv.ParseInt(string(r.Form["terminal_id"][0]), 10, 32)
	if err != nil {
		log.Println("terminal id convert int error", err, r.Form["terminal_id"][0])
		return nil, -12
	}
	return &QueryParams{
		SlotId: byte(slot_id),
		DeviceId: uint32(device_id),
		ActionId: uint32(action_id),
		TerminalId: uint32(terminal_id),
	}, 0
}

// curl -d "" "http://127.0.0.1:12000/terminal/outstock?slot_id=1&device_id=1&action_id=2&terminal_id=1"
func OutStockHandler(w http.ResponseWriter, r *http.Request, args []string, pd []byte) {
	queryParams, err := checkParams(r)
	if err != 0 {
		response_err(err, &w)
		return
	}

	log.Info("queryParams", queryParams)
	t := GetTerminalById(queryParams.TerminalId)
	if t == nil {
		response_err(-13, &w)
		return
	}
	t.SendOutStockMessage(queryParams.ActionId, queryParams.TerminalId, queryParams.SlotId)
	//sqlutils.OutStockTerminalDeviceId(uint32(queryParams.SlotId), queryParams.TerminalId)
	//testInsert(queryParams.SlotId, queryParams.DeviceId, queryParams.ActionId, queryParams.TerminalId, false)
	data := map[string]interface{}{
		"err": 0,
	}
	res, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	io.WriteString(w, string(res))
}

//curl -d "" "http://127.0.0.1:12000/terminal/instock?slot_id=1&device_id=1&action_id=2&terminal_id=1"
func InStockHandler(w http.ResponseWriter, r *http.Request, args []string, pd []byte) {
	queryParams, err := checkParams(r)
	if err != 0 {
		response_err(err, &w)
		return
	}

	log.Info("queryParams", queryParams)
	t := GetTerminalById(queryParams.TerminalId)
	if t == nil {
		response_err(-13, &w)
		return
	}
	t.SendInStockMessage(queryParams.ActionId, queryParams.TerminalId, queryParams.SlotId)
	//sqlutils.InStockTerminalDeviceId(queryParams.DeviceId, uint32(queryParams.SlotId), queryParams.TerminalId)
	//testInsert(queryParams.SlotId, queryParams.DeviceId, queryParams.ActionId, queryParams.TerminalId, true)
	data := map[string]interface{}{
		"err": 0,

	}
	res, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	io.WriteString(w, string(res))
}

// for test
func PackOutStockConfirmEventData(sid byte, did, aid uint32) []byte {
	ret := make([]byte, 0, 6)
	slotId := byte(sid)
	deviceId := message.IntToBytes4(did)
	result := byte(0)
	ret = append(ret, slotId)
	ret = append(ret, deviceId...)
	ret = append(ret, result)

	return ret
	//ParseEventData(1, ret)

}

func testInsert(sid byte, did, aid, tid uint32, f bool){
	var sequence, terminalId uint32
	sequence = uint32(time.Now().Unix())
	terminalId = tid
	createTime := uint32(time.Now().Unix())
	eventData := PackOutStockConfirmEventData(sid, did, aid)
	eventDetail := &message.EventDetail{
		SlotId: int32(sid),
		DeviceId: did,
		Result: 0,
	}
	event := message.OutStockConfirm
	if f{
		event = message.InStockConfirm
	}
	m := &message.Message{
		Version:    1,
		Sequence:   sequence,
		Direction:  1,
		Event:      byte(event),
		TerminalId: terminalId,
		CreateTime: createTime,
		EventData:  eventData,
	}
	res := m.Pack()
	m.InsertMessage(eventDetail, res)

}

// for test