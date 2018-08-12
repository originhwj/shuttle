package main

import (
	"net/http"
	"encoding/json"
	"io"
	"strconv"
	"time"
	"./utils/message"
	"./utils/sqlutils"
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



func OutStockHandler(w http.ResponseWriter, r *http.Request, args []string, pd []byte) {
	_, exist := r.Form["slot_id"]
	if !exist {
		log.Println("slot_id not in request")
		response_err(-5, &w)
		return
	}
	slot_id, err := strconv.ParseInt(string(r.Form["slot_id"][0]), 10, 32)
	if err != nil {
		log.Println("slot id convert int error", err, r.Form["slot_id"][0])
		response_err(-6, &w)
		return
	}

	_, exist = r.Form["device_id"]
	if !exist {
		log.Println("device_id not in request")
		response_err(-7, &w)
		return
	}
	device_id, err := strconv.ParseInt(string(r.Form["device_id"][0]), 10, 32)
	if err != nil {
		log.Println("device id convert int error", err, r.Form["device_id"][0])
		response_err(-8, &w)
		return
	}

	_, exist = r.Form["action_id"]
	if !exist {
		log.Println("action_id not in request")
		response_err(-9, &w)
		return
	}
	action_id, err := strconv.ParseInt(string(r.Form["action_id"][0]), 10, 32)
	if err != nil {
		log.Println("action id convert int error", err, r.Form["action_id"][0])
		response_err(-10, &w)
		return
	}

	_, exist = r.Form["terminal_id"]
	if !exist {
		log.Println("terminal_id not in request")
		response_err(-11, &w)
		return
	}
	terminal_id, err := strconv.ParseInt(string(r.Form["terminal_id"][0]), 10, 32)
	if err != nil {
		log.Println("terminal id convert int error", err, r.Form["terminal_id"][0])
		response_err(-12, &w)
		return
	}

	log.Info(slot_id, device_id, action_id, terminal_id)
	sqlutils.OutStockTerminalDeviceId(int32(slot_id), int32(terminal_id))
	testInsert(byte(slot_id), int32(device_id), int32(action_id), int32(terminal_id), false)
	data := map[string]interface{}{
		"err": 0,

	}
	res, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	io.WriteString(w, string(res))
}

func InStockHandler(w http.ResponseWriter, r *http.Request, args []string, pd []byte) {
	_, exist := r.Form["slot_id"]
	if !exist {
		log.Println("slot_id not in request")
		response_err(-5, &w)
		return
	}
	slot_id, err := strconv.ParseInt(string(r.Form["slot_id"][0]), 10, 32)
	if err != nil {
		log.Println("slot id convert int error", err, r.Form["slot_id"][0])
		response_err(-6, &w)
		return
	}

	_, exist = r.Form["device_id"]
	if !exist {
		log.Println("device_id not in request")
		response_err(-7, &w)
		return
	}
	device_id, err := strconv.ParseInt(string(r.Form["device_id"][0]), 10, 32)
	if err != nil {
		log.Println("device id convert int error", err, r.Form["device_id"][0])
		response_err(-8, &w)
		return
	}

	_, exist = r.Form["action_id"]
	if !exist {
		log.Println("action_id not in request")
		response_err(-9, &w)
		return
	}
	action_id, err := strconv.ParseInt(string(r.Form["action_id"][0]), 10, 32)
	if err != nil {
		log.Println("action id convert int error", err, r.Form["action_id"][0])
		response_err(-10, &w)
		return
	}

	_, exist = r.Form["terminal_id"]
	if !exist {
		log.Println("terminal_id not in request")
		response_err(-11, &w)
		return
	}
	terminal_id, err := strconv.ParseInt(string(r.Form["terminal_id"][0]), 10, 32)
	if err != nil {
		log.Println("terminal id convert int error", err, r.Form["terminal_id"][0])
		response_err(-12, &w)
		return
	}

	log.Info(slot_id, device_id, action_id, terminal_id)
	sqlutils.InStockTerminalDeviceId(int32(device_id), int32(slot_id), int32(terminal_id))
	testInsert(byte(slot_id), int32(device_id), int32(action_id), int32(terminal_id), true)
	data := map[string]interface{}{
		"err": 0,

	}
	res, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	io.WriteString(w, string(res))
}

// for test
func PackOutStockConfirmEventData(sid byte, did, aid int32) []byte {
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

func testInsert(sid byte, did, aid, tid int32, f bool){
	var sequence, terminalId int32
	sequence = int32(time.Now().Unix())
	terminalId = tid
	createTime := int32(time.Now().Unix())
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