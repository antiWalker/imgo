package main

import (
	"flag"
	"imgo/libs"
	"runtime"

	"github.com/spf13/viper"
	"time"
)

var (
	Conf     *Config
	confPath string
)

func init() {
	flag.StringVar(&confPath, "d", "../config/", " set worker config file path")
}

type Config struct {
	Base        BaseConf       `mapstructure:"base"`
	Websocket   WebsocketConf  `mapstructure:"websocket"`
	RpcPushAdds []RpcPushAddrs `mapstructure:"rpcPushAddrs"`
	EtcdInfo    Etcd           `mapstructure:"etcd"`
}

type Etcd struct {
	Host            []string `mapstructure:"host"`
	BasePath        string `mapstructure:"basePath"`
	ServerId        string `mapstructure:"ServerId"`
	ServerPathWorker string `mapstructure:"serverPathWorker"`
}

type RpcPushAddrs struct {
	Key  int8   `mapstructure:"key"`
	Addr string `mapstructure:"addr"`
}
type WebsocketConf struct {
	Bind string `mapstructure:"bind"` //
	BeatingInterval int `mapstructure:"beatingInterval"` //
	RetransInterval int `mapstructure:"retransInterval"` //
}

// 基础的配置信息
type BaseConf struct {
	Pidfile         string   `mapstructure:"pidfile"`
	ServerId        string   `mapstructure:"serverId"`
	MaxProc         int
	PprofBind       []string `mapstructure:"pprofBind"` // 性能监控的域名端口
	Logfile         string   `mapstructure:"logfile"`   // log 文件
	WriteWait       time.Duration
	PongWait        time.Duration
	PingPeriod      time.Duration
	MaxMessageSize  int64
	BroadcastSize   int
	SignalSize		int
	ReadBufferSize  int
	WriteBufferSize int
	RedisAddr       string `mapstructure:"RedisAddr"` //
	RedisPw         string `mapstructure:"redisPw"`
	RedisDefaultDB  int    `mapstructure:"redisDefaultDB"`
	RedisPoolSize   int
	RedisKeyTtl     int
}

func InitConfig() (err error) {
	Conf = NewConfig()
	viper.SetConfigName("worker")
	viper.SetConfigType("toml")
	viper.AddConfigPath(confPath)

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	if err := viper.Unmarshal(&Conf); err != nil {
		libs.ZapLogger.Panic("unable to decode into struct："+err.Error())
	}

	return nil
}

func NewConfig() *Config {
	return &Config{
		Base: BaseConf{
			Pidfile:         "/tmp/worker.pid",
			Logfile:         "/worker.log",
			MaxProc:         runtime.NumCPU(),
			PprofBind:       []string{"localhost:7777"},
			WriteWait:       10 * time.Second,
			PongWait:        60 * time.Second,
			PingPeriod:      20 * time.Second,
			MaxMessageSize:  512,
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			BroadcastSize:   256,
			SignalSize:		 10,
		},
		Websocket: WebsocketConf{
			Bind: ":8989",
		},
	}
}
