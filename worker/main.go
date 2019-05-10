package main

import (
	"flag"
	"imgo/libs"
)

func main() {
	flag.Parse()
	defer libs.ZapLogger.Sync()

	if err := InitConfig(); err != nil {
		return
	}

	libs.InitLogger(Conf.Base.Logfile, "worker")
	// 设置cpu 核数
	//runtime.GOMAXPROCS(Conf.Base.MaxProc)
	// 加入性能监控
	libs.StartPprof(Conf.Base.PprofBind)

	// 加入监控 后补
	if err := InitRedis(); err != nil {
		libs.ZapLogger.Panic("InitRedis() fatal error : "+err.Error())
	}
	DefaultServer = &Server{
		WriteWait:       Conf.Base.WriteWait,
		PongWait:        Conf.Base.PongWait,
		PingPeriod:      Conf.Base.PingPeriod,
		MaxMessageSize:  Conf.Base.MaxMessageSize,
		ReadBufferSize:  Conf.Base.ReadBufferSize,
		WriteBufferSize: Conf.Base.WriteBufferSize,
		BroadcastSize:   Conf.Base.BroadcastSize,
		SignalSize:  	 Conf.Base.SignalSize,
	}

	if err := InitPushRpc(Conf.RpcPushAdds); err != nil {
		libs.ZapLogger.Fatal(err.Error())
	}

	if err := InitWebsocket(Conf.Websocket.Bind); err != nil {
		libs.ZapLogger.Fatal(err.Error())
	}

}
