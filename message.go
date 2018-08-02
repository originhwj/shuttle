package main

import (
	"bytes"
	"encoding/binary"
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
