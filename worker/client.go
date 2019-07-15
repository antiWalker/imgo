package main

import (
	"github.com/gorilla/websocket"
	"imgo/libs"
	"time"
)

type Client struct {
	Uid          string
	Uuid         string
	Time         int64
	Next         *Client
	Prev         *Client
	conn         *websocket.Conn
	IsSignIn	 bool
	cancel       chan int
	receiveHeart chan int
	startHeart   chan int
	writeClose   chan int
	broadcast    chan *libs.Proto
	receiveAck   chan int
	retansClose  chan int
}

func NewClient(svr int,signalSvr int) *Client {
	c := new(Client)
	c.broadcast = make(chan *libs.Proto, svr)
	c.Next = nil
	c.Prev = nil
	c.IsSignIn = true
	c.cancel = make(chan int, signalSvr)
	c.receiveHeart = make(chan int, signalSvr)
	c.startHeart = make(chan int, signalSvr)
	c.writeClose = make(chan int, signalSvr)
	c.receiveAck = make(chan int, signalSvr)
	c.retansClose = make(chan int, signalSvr)
	return c
}

func (cl *Client) Push(p *libs.Proto) (err error) {
	select {
	case cl.broadcast <- p:
	default:
	}

	return
}

func (cl *Client) Release() {

	// <- 1
	libs.ZapLogger.Info("released")

	if cl.IsSignIn ==true{
		cl.cancel <- 1 //退出心跳
		cl.writeClose <- 1 //关闭写
		cl.retansClose <-1 //关闭重发
		cl.IsSignIn = false
		//del local
		cl.conn.Close()
		delete(cl.Uuid)
		//
		uid := cl.Uid
		uuid := cl.Uuid
		DelHash(uid, uuid)
	}
}

// HeartBeat 服务端检测链接是否正常
func (cl *Client) HeartBeat() {
	defer func() {
		libs.ZapLogger.Info("quit heart")
		if r := recover(); r != nil {
			libs.ZapLogger.Info("err is redis may be down")
		}
	}()
	BeatDuration := time.Second * time.Duration(Conf.Websocket.BeatingInterval)
	BeatDelay := time.NewTimer(BeatDuration)
	defer BeatDelay.Stop()
	for {
		BeatDelay.Reset(BeatDuration)
		select {
		case <-BeatDelay.C://time.After(time.Second * time.Duration(Conf.Websocket.BeatingInterval)):
			libs.ZapLogger.Info("in BeatingInterval s,the heartAck is not received,so release it")
			cl.Release()
			return
		case <-cl.receiveHeart:
			libs.ZapLogger.Info("in BeatingInterval s,the heartAck is received")
			err :=UpdateUserExpire(cl.Uid)
			if err !=nil{
				libs.ZapLogger.Error("err is redis may be down")
			}
		case <-cl.startHeart:
			libs.ZapLogger.Info("from now ,start the heartCheck")
		case <-cl.cancel:
			libs.ZapLogger.Info("close the heartCheck")
			return
		}
	}

}

// 消息重新投递
func (cl *Client) Retransmission(msg string) {
	defer func() {
		if r := recover(); r != nil {
			libs.ZapLogger.Info("err is redis may be down")
		}
	}()
	var retranNumber = 0
	RetranDuration := time.Second * time.Duration(Conf.Websocket.RetransInterval)
	RetranDelay := time.NewTimer(RetranDuration)
	defer RetranDelay.Stop()
	for {
		RetranDelay.Reset(RetranDuration)
		select {
		case <-RetranDelay.C://time.After(time.Second * time.Duration(Conf.Websocket.RetransInterval)):
			if retranNumber < 3 {
				libs.ZapLogger.Info("the AckFrame is not received,so retransmission")
				cl.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				w, err := cl.conn.NextWriter(websocket.TextMessage)
				if err != nil {
					libs.ZapLogger.Error(err.Error())
					return
				}
				w.Write([]byte(msg))
				if err := w.Close(); err != nil {
					return
				}
			} else {
				//已经超过三次了。记录失败的逻辑。并断开连接.
				libs.ZapLogger.Info("retrans it more than 3 times,so release")
				cl.Release()
				return
			}
			retranNumber++
		case <-cl.receiveAck:
			libs.ZapLogger.Info("the client receive it in RetransInterval s")
			return
		case <-cl.retansClose:
			libs.ZapLogger.Info("close the retransmission")
			return
		}
	}
}
