package main

import (
	"github.com/go-redis/redis"
	"imgo/libs"
	"strconv"
	"strings"
)

const REDIS_FIELD_LEN_MIN int = 5 //Redis中标识用户连接的field的最短长度

type redisClusterConf struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type SessionInfo struct {
	ServerId int16 //RPC服务器id
	Channel  int   //渠道号
	Role     int   //用户角色
	Product  int   //业务线
}

var (
	RedisCli  *redis.Client
	resisConf redisClusterConf
)

func getRedisConf() (string, error) {
	valueRedis, err := libs.GetRedisConf("lionet", 6380)
	if err != nil {
		libs.ZapLogger.Error(err.Error())
	}
	return valueRedis, err
}

//InitRedis 初始化redis连接
//从配置文件读取redis地址、密码、数据库号
func InitRedis() (err error) {
	RedisCli = redis.NewClient(&redis.Options{
		Addr:     Conf.Base.RedisAddr,
		Password: Conf.Base.RedisPw,        // no password set
		DB:       Conf.Base.RedisDefaultDB, // use default DB
	})
	if pong, err := RedisCli.Ping().Result(); err != nil {
		libs.ZapLogger.Error("RedisCli Ping Result pong:" + string(pong) + " err:" + err.Error())
		return err
	}

	return nil
}

// 具体get set的业务封装函数
// 根据toUserid 查询归属在哪个server上
func GetUserPlace(uid string) (map[string]string, error) {
	if RedisCli == nil {
		libs.ZapLogger.Error("RedisCli == nil")
		return make(map[string]string), nil
	}
	key := uid
	value, err := RedisCli.HGetAll(libs.REDIS_PREFIX + key).Result()
	if err != nil {
		libs.ZapLogger.Error("HGetAll err key=" + key + " err=" + err.Error())
		return make(map[string]string), err
	}

	return value, err

}

func GetSessionInfo(v string) (SessionInfo, bool) {
	var info SessionInfo
	len := len(v)
	if len < REDIS_FIELD_LEN_MIN {
		libs.ZapLogger.Error("lenv < 5 v=" + v)
		return info, false
	}
	index := strings.LastIndex(v, "_")
	if index == -1 {
		libs.ZapLogger.Error("index == -1 v=" + v)
		return info, false
	}

	info.Product, _ = strconv.Atoi(v[index+1 : len])
	info.Role, _ = strconv.Atoi(v[index-1 : index])
	info.Channel, _ = strconv.Atoi(v[index-2 : index-1])
	serverid, _ := strconv.Atoi(v[:index-2])
	info.ServerId = int16(serverid)

	return info, true
}

//DelConnection 删除hash中名称为key，键为field的域
//用户设备断线后删除连接信息，逻辑服务挂了需要删除其上所有连接
func DelConnection(key string, field string) bool {
	if RedisCli == nil {
		libs.ZapLogger.Error("RedisCli == nil")
		return false
	}
	err := RedisCli.HDel(libs.REDIS_PREFIX+key, field).Err()
	if err != nil {
		libs.ZapLogger.Error("HDel err key=" + key + " field=" + field + " err=" + err.Error())
		return false
	}

	return true
}
