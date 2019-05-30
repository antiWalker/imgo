package main

import (
	"github.com/go-redis/redis"
	"github.com/json-iterator/go"
	"imgo/libs"
	"strconv"
)

type redisClusterConf struct {
	Host string `json:"host"`
	Port int    `json:"port"`
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
	redisconf, err := getRedisConf()
	if err != nil {
		libs.ZapLogger.Error(err.Error())
		return err
	}

	var jsoniterator = jsoniter.ConfigCompatibleWithStandardLibrary
	err = jsoniterator.Unmarshal([]byte(redisconf), &resisConf)
	if err != nil {
		libs.ZapLogger.Error(err.Error())
		return err
	}

	port := strconv.Itoa(resisConf.Port)
	redisAddr := resisConf.Host + ":" + port

	RedisCli = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",                       // no password set
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
