package main

import (
	"fmt"
	"io"
	"net"
	"time"

	"../utils/message"
	"../utils/sqlutils"
	"../config"
)

var (
	StartByte = []byte{0x02}
	EndByte   = []byte{0x03}
	Ping      = []byte{1}

	Host = config.Host

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
		ret = append(ret, message.IntToBytes4(uint32(i))...)
	}
	fmt.Println(len(ret), ret)
	return ret
	//ParseEventData(1, ret)

}

func PackPing() []byte {
	var sequence, terminalId uint32
	sequence = uint32(time.Now().Unix())
	terminalId = 1
	createTime := uint32(time.Now().Unix())
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
	var sequence, terminalId uint32
	sequence = uint32(time.Now().Unix())
	terminalId = 1
	createTime := uint32(time.Now().Unix())
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
			pingResponse := []byte{2,0,0,0,57, 1,91,107, 171, 222, 1, 1, 0, 0, 0, 1, 91, 107, 171, 222, 0, 0, 0, 30, 0, 0, 64, 55, 116, 239,
				136, 185, 119, 133, 64, 40, 63, 52, 214, 161, 97, 229, 1, 2, 1, 0, 0, 0, 1, 2, 0, 0, 0, 2, 0, 0, 39, 176, 3}
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
	var sequence, terminalId uint32
	sequence = uint32(time.Now().Unix())
	terminalId = 1
	createTime := uint32(time.Now().Unix())
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

func shifting(a uint32) uint32 {
	a = a << 3
	return a
}

func main() {
	//Dial()
	//a := []byte{2,0,0,0,57, 1,91,107, 171, 222, 1, 1, 0, 0, 0, 1, 91, 107, 171, 222, 0, 0, 0, 30, 0, 0, 64, 55, 116, 239,
	//	136, 185, 119, 133, 64, 40, 63, 52, 214, 161, 97, 229, 1, 2, 1, 0, 0, 0, 1, 2, 0, 0, 0, 2, 0, 0, 39, 176, 3}
	//fmt.Println(message.Bytes4ToInt(a))
	fmt.Println(shifting(3067566100))
}
