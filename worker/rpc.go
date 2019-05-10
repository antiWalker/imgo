package main

import (
	"context"
	"github.com/smallnest/rpcx/server"
	"imgo/libs"
)

type PushRpc int

func InitPushRpc(addrs []RpcPushAddrs) (err error) {
	var (
		network, addr string
	)
	for _, bind := range addrs {
		if network, addr, err = libs.ParseNetwork(bind.Addr); err != nil {
			libs.ZapLogger.Panic("InitLogicRpc ParseNetwork error : "+err.Error())
		}
		go createServer(network, addr)
	}
	return
}

func createServer(network string, addr string) {
	s := server.NewServer()
	s.RegisterName("PushRpc", new(PushRpc), "")
	s.Serve(network, addr)
}

func (rpc *PushRpc) PushMsg(ctx context.Context, args *libs.PushMsgArg, SuccessReply *libs.RpcSuccessReply) (err error) {
	defer func() {
		if r := recover(); r != nil {
			libs.ZapLogger.Error("err is  : "+err.Error())
		}
	}()

	SuccessReply.Code = libs.SuccessReply
	SuccessReply.Msg = libs.SuccessReplyMsg
	libs.ZapLogger.Info("uuid+msg is  : "+string(args.Uuid)+" "+string(args.Msg))
	Client := load(args.Uuid)
	if Client != nil {
		tob := &libs.Proto{Ver: 1, Operation: libs.OP_SINGLE_SEND, Body: string(args.Msg)}
		err = Client.Push(tob)
		return
	} else {
		//exception conn，must to del it from redis
		libs.ZapLogger.Error("local map is nil")
		uid := args.Uuid[13:]
		DelHash(uid, args.Uuid)
	}
	return
}
