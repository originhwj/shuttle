package sqlutils

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"../../config"
	"fmt"
	"../log"
	"time"
)


var(
	db *sql.DB
)

const (
	DefDateLayout     = "2006-01-02"
	DefDatetimeLayout = "2006-01-02 15:04:05"
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

func GetPackageBySequence(sequence uint32) ([]byte, uint32){
	sql := "select package, terminal_id from tbl_package where sequence=? and direction=1 limit 1"
	var p []byte
	var terminalId uint32
	err := db.QueryRow(sql, sequence).Scan(&p, &terminalId)
	if err != nil {
		log.Error("GetPackageBySequence err", err, sequence)
		return nil, terminalId
	}
	return p, terminalId
}


func UpdateLastHeartbeat(heartbeatStatus byte, terminalId uint32){
	sql := "update tbl_terminal set last_heartbeat=?, heartbeat_status=? where terminal_id=?"
	now := time.Now().Format(DefDatetimeLayout)
	_, err := db.Exec(sql, now, heartbeatStatus, terminalId)
	if err != nil {
		log.Error("UpdateLastHeartbeat", heartbeatStatus, terminalId)
	}
}