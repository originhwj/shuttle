package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

var (
	StartByte = []byte{0x02}
	EndByte   = []byte{0x03}

	Host      = ":8888"
	INBOX_LEN = 500
)

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
			inbox:        make(chan []byte, INBOX_LEN),
		}

		go terminal.Process()
		go terminal.write_loop()
	}
}

func main() {

	tcp_server()
}
