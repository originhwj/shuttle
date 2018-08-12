package main

import (
	"fmt"
	"io"
	"net"
	"time"

	"../utils/message"
	"../utils/sqlutils"
)

var (
	StartByte = []byte{0x02}
	EndByte   = []byte{0x03}
	Ping      = []byte{1}

	Host = "svr.train-wifi.com:8888"

	INBOX_LEN = 500
)

func PackOutStockConfirmEventData() []byte {
	ret := make([]byte, 0, 6)
	slotId := byte(1)
	deviceId := message.IntToBytes4(1)
	result := byte(0)
	ret = append(ret, slotId)
	ret = append(ret, deviceId...)
	ret = append(ret, result)

	return ret
	//ParseEventData(1, ret)

}

func PackPingEventData() []byte {
	ret := make([]byte, 0, 1024)
	err_code := message.Success

	ret = append(ret, err_code...)
	ret = append(ret, message.Float64ToByte(23.45678)...)
	ret = append(ret, message.Float64ToByte(12.12345)...)
	ret = append(ret, byte(1))
	ret = append(ret, byte(2))
	for i := 1; i <= 2; i++ {
		ret = append(ret, byte(i))
		ret = append(ret, message.IntToBytes4(int32(i))...)
	}
	fmt.Println(len(ret), ret)
	return ret
	//ParseEventData(1, ret)

}

func PackPing() []byte {
	var sequence, terminalId int32
	sequence = int32(time.Now().Unix())
	terminalId = 1
	createTime := int32(time.Now().Unix())
	eventData := PackPingEventData()

	m := &message.Message{
		Version:    1,
		Sequence:   sequence,
		Direction:  1,
		Event:      message.Ping,
		TerminalId: terminalId,
		CreateTime: createTime,
		EventData:  eventData,
	}
	return m.Pack()
}

func PackOutStockConfirm() []byte {
	var sequence, terminalId int32
	sequence = int32(time.Now().Unix())
	terminalId = 1
	createTime := int32(time.Now().Unix())
	eventData := PackOutStockConfirmEventData()

	m := &message.Message{
		Version:    1,
		Sequence:   sequence,
		Direction:  1,
		Event:      message.OutStockConfirm,
		TerminalId: terminalId,
		CreateTime: createTime,
		EventData:  eventData,
	}
	return m.Pack()
}

func Dial() {
	conn, err := net.Dial("tcp", Host)
	if err != nil {
		fmt.Println(err)
		return
	}
	go func() {
		for {
			time.Sleep(2 * time.Second)
			pingResponse := PackPing()
			_, err := conn.Write(pingResponse)
			if err != nil && err != io.EOF {
				fmt.Println(err)
				return
			} else if err == io.EOF {
				fmt.Println("EOF conn")
				return
			}
		}

	}()
	for {

		//conn.Write([]byte("hello"))
		data := make([]byte, 128)
		res, err := conn.Read(data)
		if err != nil && err != io.EOF {
			fmt.Println(err)
			return
		} else if err == io.EOF {
			fmt.Println("EOF conn")
			return
		}
		fmt.Println("response:", data[:res])
	}

}

func testInsert(){
	sqlutils.SetConfig("master")
	var sequence, terminalId int32
	sequence = int32(time.Now().Unix())
	terminalId = 1
	createTime := int32(time.Now().Unix())
	eventData := PackOutStockConfirmEventData()
	eventDetail := &message.EventDetail{
		SlotId: 1,
		DeviceId: 1,
		Result: 0,
	}
	m := &message.Message{
		Version:    1,
		Sequence:   sequence,
		Direction:  1,
		Event:      message.OutStockConfirm,
		TerminalId: terminalId,
		CreateTime: createTime,
		EventData:  eventData,
	}
	res := m.Pack()
	m.InsertMessage(eventDetail, res)

}

func main() {
	testInsert()
}
