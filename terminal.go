package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"./utils/message"
)

type Terminal struct {
	Conn         net.Conn
	mu           sync.Mutex
	once         sync.Once
	err          error
	br           *bufio.Reader
	bw           *bufio.Writer
	writeTimeout time.Duration
	readTimeout  time.Duration
	TerminalId   int64
	inbox        chan []byte
	closed       bool
}

func (t *Terminal) Process() {
	defer t.Close()
	for {
		t.Conn.SetReadDeadline(time.Now().Add(t.readTimeout))
		//读取数据
		buf := make([]byte, 5)
		_, err := t.Conn.Read(buf)
		if err != nil {
			log.Error(err)
			return
		}
		start := buf[:1]
		packageLen := message.Bytes4ToInt(buf[1:5])
		fmt.Println(start, packageLen)
		if start[0] != StartByte[0] {
			log.Error("start err", start)
			return
		}
		log.Info("read package len", packageLen)
		dataLen := packageLen - 4 + 1
		data_buf := make([]byte, dataLen)
		n, err := io.ReadFull(t.br, data_buf) // 把消息包读完
		//n, err :=t.Conn.Read(data_buf)
		if err != nil || n != int(dataLen) {
			log.Error("read err", err, n, dataLen)
			return
		}
		data := data_buf[:dataLen]

		log.Info("read package data", data)
		end_buf := data[dataLen-1:]
		fmt.Println("end buff", end_buf)
		if end_buf[0] != EndByte[0] {
			log.Error("end err", end_buf)
			return
		}
		if resMsg, errCode := message.Parse2Message(data[:dataLen-1], packageLen); errCode == 0 && resMsg != nil {
			//pingResponse := message.PackPing()
			t.inbox <- resMsg.Pack()
		}

	}
}

func (t *Terminal) Close() {
	t.once.Do(func() {
		log.Info("terminal close", t)
		err := t.Conn.Close()
		if err != nil {
			log.Error("close conn err", err)
		}
		t.closed = true
	})

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
			if err != nil {
				log.Error("write err", err)
				return
			}
			log.Info("server write finish", b)
			t.bw.Flush()
		case <-time.After(60 * time.Second):
			//超时60秒,没有任何心跳信息 关掉
			log.Warn("timeout close")
			return
		}

	}
}

func (t *Terminal) SendOutStockMessage() {
	m := &message.Message{
		Version:    message.Ver,
		Sequence:   1,
		Direction:  1,
		Event:      message.OutStock,
		TerminalId: 1,
	}
	t.inbox <- m.Pack()
	log.Info("SendOutStockMessage")
}

func (t *Terminal) SendInStockMessage() {
	m := &message.Message{
		Version:    message.Ver,
		Sequence:   1,
		Direction:  1,
		Event:      message.OutStock,
		TerminalId: 1,
	}
	t.inbox <- m.Pack()
	log.Info("SendInStockMessage")
}

func (t *Terminal) crontabSendStockMessage() {
	ticker := time.NewTicker(time.Second * 5)
	for range ticker.C {
		if t.closed {
			return
		}
		t.SendInStockMessage()
	}

}
