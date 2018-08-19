package message

import (
	"bytes"
	"encoding/binary"
	"math"
	"time"
	"../log"
	"../sqlutils"
	"fmt"
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

	DefDateLayout     = "2006-01-02"
	DefDatetimeLayout = "2006-01-02 15:04:05"
)

func IntToBytes4(m uint32) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, m)

	gbyte := bytesBuffer.Bytes()

	return gbyte
}

func Bytes4ToInt(b []byte) uint32 {
	xx := make([]byte, 4)
	if len(b) == 2 {
		xx = []byte{b[0], b[1], 0, 0}
	} else {
		xx = b
	}

	bytesBuffer := bytes.NewBuffer(xx)

	var x uint32
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

func shifting(a uint32) uint32 {
	a = a << 3
	return a
}

type Message struct {
	PackageLength uint32
	Version       byte
	Sequence      uint32
	Direction     byte
	Event         byte
	TerminalId    uint32
	CreateTime    uint32
	EventLength   uint32
	EventData     []byte
	PackageHash   uint32

	EvDetail   *EventDetail
}

type EventDetail struct {
	SlotId int32
	DeviceId uint32
	Result int64
	ResponseCode int32
}

func (m *Message) Pack() []byte {
	start := []byte{0x02}
	end := []byte{0x03}

	m.CreateTime = uint32(time.Now().Unix())
	m.EventLength = uint32(len(m.EventData))
	m.PackageLength = 27 + m.EventLength
	m.PackageHash = m.PackageLength + m.Sequence + m.TerminalId + m.CreateTime + uint32(m.EventLength)

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

func ParseEventData(event, direction byte, eventData []byte, m *Message) *EventDetail{
	eventDetail := &EventDetail{}
	switch event {
	case Ping:
		m.EventData = []byte{1}
		if len(eventData) < 20 {
			log.Error("ping eventdata size not enough", len(eventData))
			return nil
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
				log.Warn("slot count not equal")
			}
		}
		log.Info("parse ping", err_code, longitude, latitude, electric, slot_count, slot_detail)
		m.EventData = []byte{0}
	case OutStock:
		if len(eventData) != 1 {
			log.Error("wrong size event data", eventData)
			return nil
		}
		errCode := eventData[0]
		if errCode != 0 {
			log.Error("wrong errCode", errCode)
			eventDetail.ResponseCode = int32(errCode)
		}else {
			log.Info("outstock success")
		}
	case InStock:
		if len(eventData) != 1 {
			log.Error("wrong size event data", eventData)
			return nil
		}
		errCode := eventData[0]
		if errCode != 0 {
			log.Error("wrong errCode", errCode)
			eventDetail.ResponseCode = int32(errCode)
		}else {
			log.Info("instock success")
		}
	case OutStockConfirm:
		m.EventData = []byte{1}
		if len(eventData) != 6 {
			log.Error("OutStockconfirm size not equal", len(eventData))
			return nil
		}
		soltId := eventData[0]
		deviceId := Bytes4ToInt(eventData[1:5])
		result := eventData[5]
		eventDetail.SlotId = int32(soltId)
		eventDetail.DeviceId = deviceId
		eventDetail.Result = int64(result)
		log.Info("outstock confirm success", soltId, deviceId, result)
		m.EventData = []byte{0}
	case InStockConfirm:
		m.EventData = []byte{1}
		if len(eventData) != 6 {
			log.Error("InStockconfirm size not equal", len(eventData))
			return nil
		}
		soltId := eventData[0]
		deviceId := Bytes4ToInt(eventData[1:5])
		result := eventData[5]
		eventDetail.SlotId = int32(soltId)
		eventDetail.DeviceId = deviceId
		eventDetail.Result = int64(result)
		log.Info("instock confirm success", soltId, deviceId, result)
		m.EventData = []byte{0}
	default:
		log.Info("not exist event", event)
	}
	return eventDetail
}

func Parse2Message(data, origin []byte, packageLength uint32) (*Message, int) {
	l := len(data)
	if l < 24 { // eventData 至少一字节
		log.Error("package size not long enough")
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
		log.Error("err valid event length", validEventLength, eventLength)
		return nil, -2
	}
	eventData := data[19 : 19+eventLength]
	packageHash := Bytes4ToInt(data[l-4:])
	hashSum := uint32(packageLength) + sequence + terminalId + createTime + eventLength
	expectHash := shifting(hashSum)
	log.Info(hashSum, IntToBytes4(hashSum))
	log.Info(expectHash, IntToBytes4(expectHash))
	log.Info(packageLength, sequence, terminalId, createTime, eventLength, hashSum, expectHash, packageHash)
	if expectHash != packageHash {
		log.Error("hash valid failed", expectHash, packageHash)
		return nil, -3
	}

	log.Info(version, sequence, direction, event, terminalId, createTime, eventLength, eventData, packageHash)
	// Todo 对解析成功的包进行入库记录
	m := &Message{
		Version:    version,
		Sequence:   sequence,
		Direction:  direction,
		Event:      event,
		TerminalId: terminalId,
		CreateTime: createTime,
	}
	eventDeatil := ParseEventData(event, direction, eventData, m)
	if eventDeatil == nil{
		// 解析有问题的包,不处理
		return nil, 0
	}
	if event != Ping{
		m.InsertMessage(eventDeatil, origin)
	}
	if event == OutStock || event == InStock { //出库入库不需要回包
		return nil, 0
	}
	// Todo 回包入库
	m.Direction = 2
	if event == OutStockConfirm { // 同步slot
		m.EvDetail = eventDeatil
		sqlutils.OutStockTerminalDeviceId(uint32(eventDeatil.SlotId), m.TerminalId)
	} else if event == InStockConfirm {
		m.EvDetail = eventDeatil
		sqlutils.InStockTerminalDeviceId(eventDeatil.DeviceId, uint32(eventDeatil.SlotId), m.TerminalId)
	}
	return m, 0

}

func PackStockEventData(slotId byte) []byte {
	ret := make([]byte, 0, 1)
	ret = append(ret, slotId)
	return ret

}

func (m *Message) SelfLog() string{
	return fmt.Sprintf("%#v", m)
}

func (m *Message) InsertMessage(eventDetail *EventDetail, pack []byte){
	db := sqlutils.GetShuttleDB()
	sql := `insert into tbl_package(version, terminal_id, sequence, direction,event,send_time,create_time,
	ip, slot_id, device_id, result,response_code,package) value (?,?,?,?,?,?,?,?,?,?,?,?,?)`
	send_time :=  time.Unix(int64(m.CreateTime), 0).Format(DefDatetimeLayout)
	now := time.Now().Format(DefDatetimeLayout)
	res, err := db.Exec(sql, m.Version, m.TerminalId, m.Sequence, m.Direction, m.Event, send_time, now, "",
		eventDetail.SlotId, eventDetail.DeviceId, eventDetail.Result, eventDetail.ResponseCode, pack)
	log.Println(res, err, m.SelfLog())
	if err != nil {
		 log.Error("InsertMessage err", err, m.SelfLog())
	}
	log.Info("InsertMessage reply", res)
}
