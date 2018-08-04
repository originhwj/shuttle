package message

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

var (
	Success = []byte{0x00, 0x00}
	Fail1   = []byte{0x00, 0x01}
	Fail2   = []byte{0x00, 0x02}
	Fail3   = []byte{0x00, 0x04}
	Fail4   = []byte{0x00, 0x08}
	Fail5   = []byte{0x00, 0x10}

	Ver byte = 1
)

const (
	Ping            = 1
	OutStock        = 2
	InStock         = 3
	OutStockConfirm = 4
	InStockConfirm  = 5
)

func IntToBytes4(m int32) []byte {
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
	Version       byte
	Sequence      int32
	Direction     byte
	Event         byte
	TerminalId    int32
	CreateTime    int32
	EventLength   int32
	EventData     []byte
	PackageHash   int32
}

func (m *Message) Pack() []byte {
	start := []byte{0x02}
	end := []byte{0x03}

	m.CreateTime = int32(time.Now().Unix())
	m.EventLength = int32(len(m.EventData))
	m.PackageLength = 27 + m.EventLength
	m.PackageHash = m.PackageLength + m.Sequence + m.TerminalId + m.CreateTime + int32(m.EventLength)

	ret := make([]byte, 0, 1024)
	ret = append(ret, start...)
	ret = append(ret, IntToBytes4(m.PackageLength)...)
	ret = append(ret, m.Version)
	ret = append(ret, IntToBytes4(m.Sequence)...)
	ret = append(ret, m.Direction)
	ret = append(ret, m.Event)
	ret = append(ret, IntToBytes4(m.TerminalId)...)
	ret = append(ret, IntToBytes4(m.CreateTime)...)
	ret = append(ret, IntToBytes4(m.EventLength)...)
	ret = append(ret, m.EventData...)
	ret = append(ret, IntToBytes4(shifting(m.PackageHash))...)
	ret = append(ret, end...)

	return ret
}

func ParseEventData(event, direction byte, eventData []byte, m *Message) {
	switch event {
	case Ping:
		m.EventData = []byte{1}
		if len(eventData) < 20 {
			fmt.Println("ping eventdata size not enough", len(eventData))
			return
		}
		err_code := eventData[:2]
		longitude := ByteToFloat64(eventData[2:10])
		latitude := ByteToFloat64(eventData[10:18])
		electric := eventData[18]
		slot_count := eventData[19]
		slot_detail := []byte{}
		if int(slot_count) > 0 && len(eventData) > 20 {
			slot_detail = eventData[20:]
			if len(slot_detail)/5 != int(slot_count) {
				fmt.Println("slot count not equal")
			}
		}
		fmt.Println("parse ping", err_code, longitude, latitude, electric, slot_count, slot_detail)
		m.EventData = []byte{0}
	case OutStock:
		if len(eventData) != 1 {
			fmt.Println("wrong size event data", eventData)
			return
		}
		errCode := eventData[0]
		if errCode != 0 {
			fmt.Println("wrong errCode", errCode)
		}
		fmt.Println("outstock success")
	case InStock:
		if len(eventData) != 1 {
			fmt.Println("wrong size event data", eventData)
			return
		}
		errCode := eventData[0]
		if errCode != 0 {
			fmt.Println("wrong errCode", errCode)
		}
		fmt.Println("instock success")
	case OutStockConfirm:
		m.EventData = []byte{1}
		if len(eventData) != 6 {
			fmt.Println("OutStockconfirm size not equal", len(eventData))
			return
		}
		soltId := eventData[0]
		deviceId := Bytes4ToInt(eventData[1:5])
		result := eventData[5]
		fmt.Println("outstock confirm success", soltId, deviceId, result)
		m.EventData = []byte{0}
	case InStockConfirm:
		m.EventData = []byte{1}
		if len(eventData) != 6 {
			fmt.Println("InStockconfirm size not equal", len(eventData))
			return
		}
		soltId := eventData[0]
		deviceId := eventData[1:5]
		result := eventData[5]
		fmt.Println("instock confirm success", soltId, deviceId, result)
		m.EventData = []byte{0}
	default:
		fmt.Println("not exist event", event)
	}
}

func Parse2Message(data []byte, packageLength int32) (*Message, int) {
	l := len(data)
	if l < 24 { // eventData 至少一字节
		fmt.Println("package size not long enough")
		return nil, -1
	}
	validEventLength := l - 23
	//fmt.Println(l)
	version := data[0]
	sequence := Bytes4ToInt(data[1:5])
	//fmt.Println(sequence, data[1:5])
	direction := data[5]
	event := data[6]
	terminalId := Bytes4ToInt(data[7:11])
	createTime := Bytes4ToInt(data[11:15])
	eventLength := Bytes4ToInt(data[15:19])
	if validEventLength != int(eventLength) {
		fmt.Println("err valid event length", validEventLength, eventLength)
		return nil, -2
	}
	eventData := data[19 : 19+eventLength]
	packageHash := Bytes4ToInt(data[l-4:])

	expectHash := shifting(int32(packageLength) + sequence + terminalId + createTime + int32(eventLength))
	if expectHash != packageHash {
		fmt.Println("hash valid failed", expectHash, packageHash)
		return nil, -3
	}

	fmt.Println(version, sequence, direction, event, terminalId, createTime, eventLength, eventData, packageHash)
	m := &Message{
		Version:    version,
		Sequence:   sequence,
		Direction:  2,
		Event:      event,
		TerminalId: terminalId,
	}
	ParseEventData(event, direction, eventData, m)
	if event == OutStock || event == InStock { //出库入库不需要回包
		return nil, 0
	}
	return m, 0

}

func PackStockEventData(slotId byte) []byte {
	ret := make([]byte, 0, 1)
	ret = append(ret, slotId)
	return ret

}
