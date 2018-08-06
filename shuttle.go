package main

import (
	logger "./utils/log"
	"bufio"
	"fmt"
	"net"
	"time"
	"flag"
	"os"
	"syscall"
)

var (
	StartByte = []byte{0x02}
	EndByte   = []byte{0x03}

	Host      = ":8888"
	INBOX_LEN = 500
	env       *string
	logPath   *string
	log       *logger.Logger
)

func init_log(log_path string) {
	filename := log_path + ".log"
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		logger.Println("fail to create log file! err:", err)
		return
	}
	log = logger.New(file, "", logger.Ldate|logger.Ltime|logger.Lmicroseconds|logger.Lshortfile, logger.FWARN)
	syscall.Dup2(int(file.Fd()), 1)
	syscall.Dup2(int(file.Fd()), 2)
	logger.SetLogger(log)
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
			inbox:        make(chan []byte, INBOX_LEN),
		}

		go terminal.Process()
		go terminal.write_loop()
	}
}

func main() {

	env = flag.String("env", "test", "dev environment")
	logPath = flag.String("logPath", "./shuttle", "log path")
	flag.Parse()
	init_log(*logPath) //初始化日志
	tcp_server()
}
