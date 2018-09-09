package main

import (
	"../utils/redisutils"
	"../utils/sqlutils"
	"../utils/callback"
	"fmt"
	"strconv"

)

func testRedis()  {
	datas := redisutils.GetMessageSequence()
	for i := 0; i+1 < len(datas); i += 2 {
		value := datas[i]
		seq, err := strconv.Atoi(value)
		if err != nil {
			continue
		}
		msg, tid := sqlutils.GetPackageBySequence(uint32(seq))
		if msg != nil{
			fmt.Println(msg, tid)
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
	//testSql()
	//fmt.Println(redisutils.RemoveMessageSequenceList(1))
	testGen()
	//testCallback()
}
