package sqlutils

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"../../config"
	"fmt"
	"../log"
)


var(
	db *sql.DB
)


func GetShuttleDB() *sql.DB {
	return db
}

func SetConfig(env string){
	if env == "master" {
		db, _ = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/train_share?charset=utf8mb4", config.SQL_USER,
			config.SQL_PWD, config.SQL_HOST ))

	} else { //测试环境
		db, _ = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/train_share?charset=utf8mb4", config.SQL_USER,
			config.SQL_PWD, config.SQL_HOST ))


	}
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
}


func OutStockTerminalDeviceId(slotId, terminalId uint32){
	sql := "update tbl_terminal_slot set device_id = 0 where terminal_id=? and slot_id=?"
	_, err := db.Exec(sql, terminalId, slotId)
	if err != nil {
		log.Error("OutStockTerminalDeviceId err", err, )
	}
}

func InStockTerminalDeviceId(deviceId, slotId, terminalId uint32){
	sql := "update tbl_terminal_slot set device_id = ? where terminal_id=? and slot_id=?"
	_, err := db.Exec(sql, deviceId, terminalId, slotId)
	if err != nil {
		log.Error("InStockTerminalDeviceId err", err, )
	}
}

func GetPackageBySequence(sequence uint32) []byte{
	return []byte{}
}