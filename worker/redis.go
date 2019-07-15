package main

import (
	"github.com/go-redis/redis"
	"imgo/libs"
	"strings"
	"time"
)


type redisClusterConf struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

var (
	RedisCli  *redis.Client
)

func InitRedis() (err error) {
	RedisCli = redis.NewClient(&redis.Options{
		Addr:     Conf.Base.RedisAddr,
		Password: Conf.Base.RedisPw,        // no password set
		DB:       Conf.Base.RedisDefaultDB, // use default DB
		PoolSize: Conf.Base.RedisPoolSize,
	})
	if pong, err := RedisCli.Ping().Result(); err != nil {
		libs.ZapLogger.Error("RedisCli Ping Result pong:"+string(pong)+" err:"+err.Error())
	}

	return
}

//将用户和服务器IP的对应关系保存到redis
func SaveUserInfo(key string, uuid string,platform string,role string) (err error) {
	if RedisCli == nil {
		libs.ZapLogger.Error("RedisCli == nil")
		return err
	}
	//RedisCli.HSet(libs.REDIS_PREFIX+key, uuid, Conf.Base.RpcInnerIp+":"+Conf.Base.RpcInnerPort)
	saveValue := Conf.Base.ServerId + platform + role
	RedisCli.HSet(libs.REDIS_PREFIX+key, uuid, saveValue)
	RedisCli.Expire(libs.REDIS_PREFIX+key, time.Second * time.Duration(Conf.Base.RedisKeyTtl))//2个小时
	return
}
//在这里要做逻辑判断，一分钟内同一个ip下的客户端建立连接次数超过60，则直接砍断类似的连接。建立黑名单机制。
func CheckUserMelanism(data string)(isMelanism bool){
	if data ==""{
		return false
	}
	var key string
	var limit int
	if i := strings.Index(data, "."); i >= 0 {
		key = "imgoip_"+data
		limit = Conf.Base.BlackipNumber
	}else{
		key = "imgouid_"+data
		limit = Conf.Base.BlackuidNumber
	}

	n, _ := RedisCli.Exists(key).Result()
	if n != 0 {
		result, _ := RedisCli.Get(key).Int64()
		if result>=int64(limit) {
			//超过60次 这个ip or uid 被封禁1分钟
			RedisCli.Expire(key, 60*time.Second)
			isMelanism = true
		}else{
			RedisCli.Incr(key)
		}
	}else{
		RedisCli.Incr(key)
		RedisCli.Expire(key, 60*time.Second)
	}
	return isMelanism
}
//续租客户端状态
func UpdateUserExpire(key string) (err error) {
	if RedisCli == nil {
		libs.ZapLogger.Error("RedisCli == nil")
		return err
	}
	RedisCli.Expire(libs.REDIS_PREFIX+key, time.Second * time.Duration(Conf.Base.RedisKeyTtl))//2个小时
	return
}

//获取hash中名称为key的所有field及其对应的value
func GetUserInfo(key string) (v map[string]string, err error) {
	if RedisCli == nil {
		libs.ZapLogger.Error("RedisCli == nil")
		return nil,err
	}
	value, err := RedisCli.HGetAll(libs.REDIS_PREFIX + key).Result()
	if err != nil {
		libs.ZapLogger.Error("HGetAll err key="+key+" err="+err.Error())
		return make(map[string]string), err
	}

	return value, err
}
func DelHash(key string, field string) bool {
	if RedisCli == nil {
		libs.ZapLogger.Error("RedisCli == nil")
		return false
	}
	err := RedisCli.HDel(libs.REDIS_PREFIX+key, field).Err()
	if err != nil {
		libs.ZapLogger.Error("HDel err key="+key+" field="+field+" err="+err.Error())
		return false
	}

	return true
}
