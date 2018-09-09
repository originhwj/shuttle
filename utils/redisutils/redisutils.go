package redisutils

import (
	"../../config"
	"../log"
	"github.com/garyburd/redigo/redis"
	"time"
)

var (
	redis_pool *redis.Pool

	MessageConfirmList = "chuansuo:confirm:msg:list"
	ExpireTs int64 = 24*3600

	SequenceGenKey = "chuansuo:sequence:gen"

)

func redisConnPool(server, password string, db int) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     16,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			connectTimeOpt := redis.DialConnectTimeout(time.Second * 3)
			writeTimeOpt := redis.DialWriteTimeout(time.Second * 3)
			readTimeOpt := redis.DialReadTimeout(time.Second * 3)
			db := redis.DialDatabase(db)
			c, err := redis.Dial("tcp", server, connectTimeOpt, writeTimeOpt, readTimeOpt, db)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func GetRedisPool() *redis.Pool {
	if redis_pool == nil {
		redis_pool = redisConnPool(config.REDIS_HOST, config.REDIS_PWD, config.REDIS_DB)
	}
	return redis_pool
}

func Getkey(key string) string {

	rp := GetRedisPool()
	redis_conn := rp.Get()
	defer redis_conn.Close()
	ret, err := redis_conn.Do("GET", key)
	s, err := redis.String(ret, err)
	if err != nil && err != redis.ErrNil {
		log.Error(err)
		return ""
	}
	return s
}


func GetMessageSequence() []string{
	rp := GetRedisPool()
	redis_conn := rp.Get()
	defer redis_conn.Close()
	now := time.Now().Unix()
	reply, err := redis_conn.Do("ZRANGEBYSCORE", MessageConfirmList, now-ExpireTs, now, "withscores")
	res, err := redis.Strings(reply, err)
	if err != nil{
		log.Error("GetMessageSequence err", err)
		return []string{}
	}
	return res
}

func AddIntoMessageSequenceList(sequence uint32) bool{
	rp := GetRedisPool()
	redis_conn := rp.Get()
	defer redis_conn.Close()
	now := time.Now().Unix()
	_, err := redis_conn.Do("ZADD", MessageConfirmList, now, sequence)
	if err != nil {
		log.Error("AddIntoMessageSequenceList err", sequence, err)
		return false
	}
	return true
}

func RemoveMessageSequenceList(sequence uint32) bool{
	rp := GetRedisPool()
	redis_conn := rp.Get()
	defer redis_conn.Close()
	_, err := redis_conn.Do("ZREM", MessageConfirmList, sequence)
	if err != nil {
		log.Error("RemoveMessageSequenceList err", sequence, err)
		return false
	}
	return true
}

func SequenceGen() uint32 {
	rp := GetRedisPool()
	redis_conn := rp.Get()
	defer redis_conn.Close()
	reply, err := redis_conn.Do("INCR", SequenceGenKey)
	res, err := redis.Int(reply, err)
	if err != nil {
		log.Error("SequenceGen err", err)
		return 0
	}
	return uint32(res)
}