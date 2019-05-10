package main

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
)

var (
	Conf     *Config
	confPath string
)

func init() {
	flag.StringVar(&confPath, "d", "../config/", " set dispatcher config file path")
}

type Config struct {
	Base       BaseConf     `mapstructure:"base"`
	WorkerConf []WorkerConf `mapstructure:"workerAddrs"`
}

// 基础的配置信息
type BaseConf struct {
	Pidfile        string `mapstructure:"pidfile"`
	MaxProc        int
	PprofAddrs     []string `mapstructure:"pprofBind"` //
	HostAddr       string   `mapstructure:"hostAddr"`  //
	RedisAddr      string   `mapstructure:"redisAddr"` //
	RedisPw        string   `mapstructure:"redisPw"`
	RedisDefaultDB int      `mapstructure:"redisDefaultDB"`
	IsDebug        bool
	UsePool        int      `mapstructure:"usePool"`  //1:使用协程池处理http请求
	PoolSize       int      `mapstructure:"poolSize"` //协程池中协程数量
}

func InitConfig() (err error) {
	viper.SetConfigName("dispatcher")
	viper.SetConfigType("toml")
	viper.AddConfigPath(confPath)

	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	if err := viper.Unmarshal(&Conf); err != nil {
		panic(fmt.Errorf("unable to decode into struct：  %s \n", err))
	}

	return nil
}
