package sqlutils

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)


var(
	db *sql.DB
)


func GetLiveRoomDB() *sql.DB {
	return db
}

func SetConfig(env string){
	if env == "master" {
		db, _ = sql.Open("mysql", "user:password(127.0.0.1:3306)/shuttle?charset=utf8mb4")

	} else { //测试环境
		test_name := "user:password(127.0.0.1:3306)/shuttle" + env + "?charset=utf8mb4"
		db, _ = sql.Open("mysql", test_name)

	}
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
}
