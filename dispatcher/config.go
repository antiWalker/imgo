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
	EtcdInfo   Etcd         `mapstructure:"etcd"`
}

type Etcd struct {
	Host                 []string `mapstructure:"host"`
	BasePath             string `mapstructure:"basePath"`
	ServerPathDispatcher string `mapstructure:"serverPathDispatcher"`
}

// 基础的配置信息
type BaseConf struct {
	Pidfile        string `mapstructure:"pidfile"`
	Logfile        string   `mapstructure:"logfile"`   // log 文件
	LogLevel	   int		`mapstructure:"logLevel"`
	MaxProc        int
	PprofAddrs     []string `mapstructure:"pprofBind"` //
	HostAddr       string   `mapstructure:"hostAddr"`  //
	RedisAddr      string   `mapstructure:"redisAddr"` //
	RedisPw        string   `mapstructure:"redisPw"`
	RedisDefaultDB int      `mapstructure:"redisDefaultDB"`
	IsDebug        bool
	UsePool        int `mapstructure:"usePool"`      //是否使用协程池处理http请求，1为使用，0为不使用
	PoolSize       int `mapstructure:"poolSize"`     //协程池中协程数量
	NoDBStrategy   int `mapstructure:"noDBStrategy"` //数据库连接断开时的策略，0为直接返回错误，1为轮询worker保证送达
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
