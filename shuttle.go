package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"time"
)

var (
	StartByte = []byte{0x02}
	EndByte   = []byte{0x03}
	Ping      = []byte{1}

	//Host = "47.96.226.207:8888"
	Host = ":8888"

	INBOX_LEN = 500
)


func PackPing() []byte {
	var sequence, terminalId int32
	version := []byte{1}
	sequence = 123456789
	event := []byte{1}
	terminalId = 1
	createTime := int32(time.Now().Unix())
	eventData := []byte{0}
	eventLength := len(eventData)
	packageLength := 26 + eventLength
	packageHash := int32(packageLength) + sequence + terminalId + createTime + int32(eventLength)

	m := &Message{
		PackageLength: int32(26 + eventLength),
		Version:       version,
		Sequence:      sequence,
		Event:         event,
		TerminalId:    terminalId,
		CreateTime:    createTime,
		EventData:     eventData,
		EventLength:   int32(eventLength),
		PackageHash:   packageHash,
	}
	return m.Pack()
}

func tcp_server() {
	var err error
	listener, err := net.Listen("tcp", ":8888")
	if err != nil {
		fmt.Println("listen err", err)
		return
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("accept err", err)
			return
		}
		terminal := Terminal{
			Conn:         conn,
			bw:           bufio.NewWriter(conn),
			br:           bufio.NewReader(conn),
			readTimeout:  10 * time.Second,
			writeTimeout: 10 * time.Second,
			inbox:     make(chan []byte, INBOX_LEN),
		}

		go terminal.Process()
		go terminal.write_loop()
		//process(conn)
	}
}

func process(conn net.Conn) {
	for {
		data := make([]byte, 128)
		//读取数据
		res, err := conn.Read(data)
		if err != nil && err != io.EOF {
			fmt.Println(err)
			return
		} else if err == io.EOF {
			fmt.Println("EOF conn")
			return
		}
		str := string(data[:res])
		fmt.Println(str)

		pingResponse := PackPing()
		conn.Write(pingResponse)

	}
}

func Dial() {
	conn, err := net.Dial("tcp", Host)
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		time.Sleep(2 * time.Second)
		pingResponse := PackPing()
		conn.Write(pingResponse)
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

func main() {

	//ret := PackPing()
	//fmt.Println(ret)
	go Dial()
	tcp_server()
}
