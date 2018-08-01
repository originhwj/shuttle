package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Terminal struct {
	Conn         net.Conn
	mu           sync.Mutex
	err          error
	br           *bufio.Reader
	bw           *bufio.Writer
	writeTimeout time.Duration
	readTimeout  time.Duration
	TerminalId   int64
	inbox   chan []byte
}

func (t *Terminal) Process() {
	for {
		t.Conn.SetReadDeadline(time.Now().Add(t.readTimeout))
		//读取数据
		buf := make([]byte, 5)
		_, err := t.Conn.Read(buf)
		if err != nil {
			fmt.Println(err)
			return
		}
		start := buf[:1]
		packageLen := Bytes4ToInt(buf[1:5])
		fmt.Println(start, packageLen)
		if start[0] != StartByte[0] {
			fmt.Println("start err", start)
			return
		}
		fmt.Println("read package len")
		dataLen := packageLen - 4 + 1
		data_buf := make([]byte, dataLen)
		n, err := io.ReadFull(t.br, data_buf) // 把消息包读完
		//n, err :=t.Conn.Read(data_buf)
		if err != nil || n != int(dataLen) {
			fmt.Println("read err", err, n, dataLen)
			return
		}
		data := data_buf[:dataLen]

		fmt.Println("data", data)
		end_buf := data[dataLen-1:]
		fmt.Println("end buff", end_buf)
		if end_buf[0] != EndByte[0] {
			fmt.Println("end err", end_buf)
			return
		}
		t.Parse2Message(data[:dataLen-1], packageLen)

		pingResponse := PackPing()
		t.inbox <- pingResponse
		//_, err = t.bw.Write(pingResponse)
		//if err != nil{
		//	fmt.Println(err)
		//}
		//t.bw.Flush()

	}
}


func (t *Terminal) Parse2Message(data []byte, packageLength int32){
	l := len(data)
	if l < 23{ // eventData 至少一字节
		return
	}
	validEventLength := l - 22
	fmt.Println(l)
	version := data[0]
	sequence := Bytes4ToInt(data[1:5])
	fmt.Println(sequence, data[1:5])
	event := data[5]
	terminalId := Bytes4ToInt(data[6:10])
	createTime := Bytes4ToInt(data[10:14])
	eventLength := Bytes4ToInt(data[14:18])
	if validEventLength != int(eventLength){
		fmt.Println("err valid event length", validEventLength, eventLength)
	}
	eventData := data[18:18+eventLength]
	packageHash := Bytes4ToInt(data[l-4:])

	expectHash := shifting(int32(packageLength) + sequence + terminalId + createTime + int32(eventLength))
	if expectHash != packageHash{
		fmt.Println("hash valid failed", expectHash, packageHash)
	}

	fmt.Println(version, sequence, event, terminalId, createTime, eventLength, eventData, packageHash)




}

func (t *Terminal) Close(){
	err := t.Conn.Close()
	if err != nil{
		fmt.Println("close conn err", err)
	}
}

func (t *Terminal) write_loop() {
	defer t.Close()
	for {
		select {
		case b := <-t.inbox:
			if b == nil {
				continue
			}
			t.Conn.SetWriteDeadline(time.Now().Add(t.writeTimeout))
			_, err := t.bw.Write(b)
			if err != nil{
				fmt.Println("write err", err)
				return
			}
			fmt.Println("server write finish", b)
			t.bw.Flush()
		case <-time.After(5 * time.Second):
		//超时60秒,没有任何心跳信息 关掉
			fmt.Println("timeout close")
			return
		}




	}
}
