package main

import (
	"math"
	"bytes"
	"encoding/binary"
	"fmt"
)

var(
	Success = []byte{0x00, 0x00}
	Fail1 = []byte{0x00, 0x01}
	Fail2 = []byte{0x00, 0x02}
	Fail3 = []byte{0x00, 0x04}
	Fail4 = []byte{0x00, 0x08}
	Fail5 = []byte{0x00, 0x10}


)

func intToBytes4(m int32) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, m)

	gbyte := bytesBuffer.Bytes()

	return gbyte
}

func Bytes4ToInt(b []byte) int32 {
	xx := make([]byte, 4)
	if len(b) == 2 {
		xx = []byte{b[0], b[1], 0, 0}
	} else {
		xx = b
	}

	bytesBuffer := bytes.NewBuffer(xx)

	var x int32
	binary.Read(bytesBuffer, binary.BigEndian, &x)

	return x
}


func Float64ToByte(float float64) []byte {
	bits := math.Float64bits(float)
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, bits)
	return bytes
}

func ByteToFloat64(bytes []byte) float64 {
	bits := binary.BigEndian.Uint64(bytes)
	return math.Float64frombits(bits)
}

func shifting(a int32) int32 {
	a = a << 3
	return a
}

type Message struct {
	PackageLength int32
	Version       []byte
	Sequence      int32
	Direction     byte
	Event         []byte
	TerminalId    int32
	CreateTime    int32
	EventLength   int32
	EventData     []byte
	PackageHash   int32
}

func (m *Message) Pack() []byte {
	start := []byte{0x02}
	end := []byte{0x03}

	ret := make([]byte, 0, 1024)
	ret = append(ret, start...)
	ret = append(ret, intToBytes4(m.PackageLength)...)
	ret = append(ret, m.Version...)
	ret = append(ret, intToBytes4(m.Sequence)...)
	ret = append(ret, m.Direction)
	ret = append(ret, m.Event...)
	ret = append(ret, intToBytes4(m.TerminalId)...)
	ret = append(ret, intToBytes4(m.CreateTime)...)
	ret = append(ret, intToBytes4(m.EventLength)...)
	ret = append(ret, m.EventData...)
	ret = append(ret, intToBytes4(shifting(m.PackageHash))...)
	ret = append(ret, end...)

	return ret
}


func ParseEventData(event byte, eventData []byte){
	switch event {
	case 1:
		err_code := eventData[:2]
		longitude := ByteToFloat64(eventData[2:10])
		latitude := ByteToFloat64(eventData[10:18])
		electric := eventData[18]
		slot_count := eventData[19]
		slot_detail := eventData[20:]
		fmt.Println("parse ping", err_code, longitude, latitude, electric, slot_count, slot_detail)
	default:
		fmt.Println("not exist event", event)
	}
}

func testPackEventData(){
	ret := make([]byte, 0, 1024)
	err_code := Success

	ret = append(ret, err_code...)
	ret = append(ret, Float64ToByte(23.45678)...)
	ret = append(ret, Float64ToByte(12.12345)...)
	ret = append(ret, byte(1))
	ret = append(ret, byte(2))
	for i:=1;i<=2;i++{
		ret = append(ret, byte(i))
		ret = append(ret, intToBytes4(int32(i))...)
	}
	fmt.Println(len(ret), ret)

	ParseEventData(1, ret)

}

