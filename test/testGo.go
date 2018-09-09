package main

import (
	"../utils/redisutils"
	"../utils/sqlutils"
	"../utils/callback"
	"fmt"
	"strconv"

	"strings"
)

func testRedis()  {
	fmt.Println(redisutils.AddIntoMessageSequenceList(1, 1, 1))
	fmt.Println(redisutils.AddIntoMessageSequenceList(1, 1, 2))
	datas := redisutils.GetMessageSequence()
	for i := 0; i+1 < len(datas); i += 2 {
		value := datas[i]
		items := strings.Split(value, "|")
		fmt.Println(items)
		if len(items) < 2{
			continue
		}

		seq, err := strconv.Atoi(items[0])
		if err != nil {
			continue
		}
		terminal_id, err := strconv.Atoi(items[1])
		if err != nil {
			continue
		}
		msg := sqlutils.GetPackageByTerminalSequence(uint32(seq), uint32(terminal_id))
		if msg != nil{
			fmt.Println(msg)
		}
	}
}


func testSql()  {
	sqlutils.SetConfig("master")
	//fmt.Println(sqlutils.GetPackageBySequence(1534078176))
}

func testCallback(){
	callback.OutStockCallBack(1,1,1,0,1)
}

func testGen(){
	fmt.Println(redisutils.SequenceGen())
}

func main() {
	//fmt.Println(redisutils.AddIntoMessageSequenceList(1))
	testSql()
	//fmt.Println(redisutils.RemoveMessageSequenceList(1))
	//testGen()
	//testCallback()
	testRedis()
}
