package main

import (
	"context"
	"github.com/smallnest/rpcx/client"
	"imgo/libs"
)

type WorkerConf struct {
	Key  int16  `mapstructure:"key"`
	Addr string `mapstructure:"addr"`
}

var (
	RpcClientList map[int16]client.XClient
)

func InitRpcConnect(WorkerConf []WorkerConf) (err error) {
	LogicAddrs := make([]*client.KVPair, len(WorkerConf))
	RpcClientList = make(map[int16]client.XClient)

	for i, bind := range WorkerConf {
		// log.Infof("bind key %d", bind.Key)
		b := new(client.KVPair)
		b.Key = bind.Addr
		// 需要转int 类型
		LogicAddrs[i] = b
		d := client.NewPeer2PeerDiscovery(bind.Addr, "")

		RpcClientList[bind.Key] = client.NewXClient(libs.RpcPushServerPath, client.Failtry, client.RandomSelect, d, client.DefaultOption)

		//log.Infof("RpcClientList addr %s, v %v ,key is %s", bind.Addr, RpcClientList[bind.Key], bind.Key)
		libs.ZapLogger.Info("RpcClientList addr " + bind.Addr + " key is  " + string(bind.Key))
	}

	return nil
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
