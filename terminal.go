package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"./utils/message"
	"./utils/redisutils"
	"./utils/sqlutils"
	"strconv"
)

var allTerminal = SafeTerminalMap{t: make(map[uint32]*Terminal)}

type SafeTerminalMap struct {
	t  map[uint32]*Terminal
	mu sync.RWMutex
}

type Terminal struct {
	Conn         net.Conn
	mu           sync.Mutex
	once         sync.Once
	err          error
	br           *bufio.Reader
	bw           *bufio.Writer
	writeTimeout time.Duration
	readTimeout  time.Duration
	TerminalId   uint32
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
		origin_data := append(buf, data...)
		if resMsg, errCode := message.Parse2Message(data[:dataLen-1], origin_data, packageLen); errCode == 0 && resMsg != nil {
			//pingResponse := message.PackPing()
			if t.TerminalId == 0 { // 第一次收到消息,注册到内存
				terminalId := resMsg.TerminalId
				allTerminal.mu.RLock()
				_, exist := allTerminal.t[terminalId]
				allTerminal.mu.RUnlock()
				if !exist {
					allTerminal.mu.Lock()
					_, exist = allTerminal.t[terminalId]
					if !exist {
						t.TerminalId = terminalId
						allTerminal.t[terminalId] = t
						log.Info("add terminal map", t.SelfLog())
					}
					allTerminal.mu.Unlock()
				}
			}
			_msg := resMsg.Pack()
			if resMsg.Event != message.Ping {
				resMsg.InsertMessage(resMsg.EvDetail, _msg)
			} else if resMsg.Event == message.InStock || resMsg.Event == message.OutStock {
				//redisutils.AddIntoMessageSequenceList(resMsg.Sequence)
			}
			t.inbox <- _msg
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
		if t.TerminalId != 0 {
			allTerminal.mu.Lock()
			delete(allTerminal.t, t.TerminalId)
			allTerminal.mu.Unlock()
			log.Info("release terminal map", t.SelfLog())
			// 更新心跳表状态
			sqlutils.UpdateLastHeartbeat(1, t.TerminalId)
		}
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

func (t *Terminal) SendOutStockMessage(actionId, terminalId uint32, slotId byte) int64{
	seq := redisutils.SequenceGen()
	if seq == 0 {
		return -14
	}
	m := &message.Message{
		Version:    message.Ver,
		Sequence:   seq,
		Direction:  1,
		Event:      message.OutStock,
		TerminalId: terminalId,
		EventData:  message.PackStockEventData(slotId, actionId),
	}
	eventDetail := &message.EventDetail{
		SlotId: int32(slotId),
		ActionId: actionId,
	}
	_msg := m.Pack()
	m.InsertMessage(eventDetail, _msg)
	t.inbox <- _msg
	log.Info("SendOutStockMessage")
	return 0
}

func (t *Terminal) SendInStockMessage(actionId, terminalId uint32, slotId byte) int64{
	seq := redisutils.SequenceGen()
	if seq == 0 {
		return -14
	}
	m := &message.Message{
		Version:    message.Ver,
		Sequence:   seq,
		Direction:  1,
		Event:      message.InStock,
		TerminalId: terminalId,
		EventData:  message.PackStockEventData(slotId, actionId),
	}
	eventDetail := &message.EventDetail{
		SlotId: int32(slotId),
		ActionId: actionId,
	}
	_msg := m.Pack()
	m.InsertMessage(eventDetail, _msg)
	t.inbox <- _msg
	log.Info("SendInStockMessage")
	return 0
}

func (t *Terminal) crontabSendStockMessage() {
	ticker := time.NewTicker(time.Second * 5)
	for range ticker.C {
		if t.closed {
			return
		}
		t.SendInStockMessage(1, 1, 1)
	}

}

func (t *Terminal) SelfLog() string {
	return fmt.Sprintf("%#v", t)
}

func GetTerminalById(terminalId uint32) *Terminal {
	allTerminal.mu.RLock()
	terminal, exist := allTerminal.t[terminalId]
	allTerminal.mu.RUnlock()
	if !exist {
		return nil
	}
	return terminal
}

//一个对redis未完成消息的消费队列
func CheckAndReSendMessage() {
	ticker := time.NewTicker(time.Second * 10)
	for range ticker.C {
		msgs := redisutils.GetMessageSequence()
		log.Info("CheckAndReSendMessage", msgs)
		for i := 0; i+1 < len(msgs); i += 2 {
			value := msgs[i]
			seq, err := strconv.Atoi(value)
			if err != nil {
				continue
			}
			msg, terminalId := sqlutils.GetPackageBySequence(uint32(seq))
			if msg != nil {
				go ResendToTerminal(msg, terminalId)
			}
		}
	}
}


func ResendToTerminal(msg []byte, terminalId uint32) {
	t := GetTerminalById(terminalId)
	if t == nil {
		log.Warn("not exist terminal, msg cannot resend", terminalId, msg)
		return
	}
	log.Info("msg resend", terminalId, msg)
	t.inbox <- msg
}