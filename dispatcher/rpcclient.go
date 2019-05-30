package main

import (
	"context"
	"github.com/smallnest/rpcx/client"
	"imgo/libs"
	"strconv"
	"strings"
	"time"
)

type WorkerConf struct {
	Key  int16  `mapstructure:"key"`
	Addr string `mapstructure:"addr"`
}

var (
	RpcClientList map[int16]client.XClient
)
func InitRpcConnect() (err error) {
	d := client.NewEtcdDiscovery(Conf.EtcdInfo.BasePath, Conf.EtcdInfo.ServerPathDispatcher, Conf.EtcdInfo.Host, nil)
	RpcClientList = make(map[int16]client.XClient, len(d.GetServices()))
	option := client.DefaultOption
	option.GenBreaker = func() client.Breaker { return client.NewConsecCircuitBreaker(5, 30*time.Second) }
	option.Retries = 10
	for _, rpcConf := range d.GetServices() {
		rpcConf.Value = strings.Replace(rpcConf.Value, "=&tps=0", "", 1)
		serverId, error := strconv.ParseInt(rpcConf.Value, 10, 8)
		if error != nil {
			libs.ZapLogger.Error("error: " + error.Error())
		}
		d := client.NewPeer2PeerDiscovery(rpcConf.Key, "")
		RpcClientList[int16(serverId)] = client.NewXClient(Conf.EtcdInfo.ServerPathDispatcher, client.Failtry, client.RandomSelect, d, option)
	}
	return
}


func PushSingleToWorker(RpcClient client.XClient, uuid string, msg string) {
	args := &libs.PushMsgArg{
		uuid,
		msg,
	}
	reply := &libs.RpcSuccessReply{
		Code: 0,
		Msg:  "",
	}
	err := RpcClient.Call(context.Background(), "PushMsg", args, reply)
	if err != nil {
		libs.ZapLogger.Error("PushSingleToWorker Call err " + err.Error())
	}
	libs.ZapLogger.Info("reply is " + reply.Msg)

}
