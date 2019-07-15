package main

import (
	"github.com/gorilla/websocket"
	"github.com/json-iterator/go"
	"imgo/libs"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	DefaultServer *Server
)
/*
//Test for examples
func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path[1:])
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "../demo/index.html")
}
*/
func InitWebsocket(bind string) (err error) {
	//http.HandleFunc("/", serveHome)

	http.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
		serveWs(DefaultServer, w, r)
	})
	err = http.ListenAndServe(bind, nil)
	return err
}

// serveWs handles websocket requests from the peer.
func serveWs(server *Server, w http.ResponseWriter, r *http.Request) {
	mfwUid, err := r.Cookie("mfw_uid")
	//ipPort := r.RemoteAddr
	ipPort := r.Header.Get("X-Real-IP")
	if ipPort[:5] == "10.3."{
		ipInfo :=r.Header.Get("X-Forwarded-For")
		XForwardedFor :=strings.Split(ipInfo,",")
		ipPort = XForwardedFor[0]
		/*
		libs.ZapLogger.Info("-----1---")
		ipPort5 := r.Header
		for k := range ipPort5 {
			value := r.Header.Get(k)
			libs.ZapLogger.Info("key==>"+k+" value==>"+value)
		}
		libs.ZapLogger.Info("-----2---")
		*/
	}
	libs.ZapLogger.Info("ip is "+ipPort)
	var cl *Client
	if err == nil && mfwUid.Value !="" {

		libs.ZapLogger.Info(mfwUid.Value+" login")
		var upgrader = websocket.Upgrader{
			ReadBufferSize:  DefaultServer.ReadBufferSize,
			WriteBufferSize: DefaultServer.WriteBufferSize,
		}
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }

		conn, err := upgrader.Upgrade(w, r, nil)

		if err != nil {
			libs.ZapLogger.Error(err.Error())
			return
		}
		from, err := r.Cookie("f")
		var platform string
		if err == nil && from.Value !="" {
			platform = from.Value
		} else {
			libs.ZapLogger.Error("have no from,so quit now")
			return
		}
		//role just
		roleInfo, err := r.Cookie("r")
		var role string
		if err == nil && roleInfo.Value !="" {
			role = roleInfo.Value
		} else {
			libs.ZapLogger.Error("have no role,so quit now")
			return
		}
		/*
		productInfo, err := r.Cookie("p")
		var product string
		if err == nil && productInfo.Value !="" {
			product = productInfo.Value
		} else {
			libs.ZapLogger.Info("have no product,so quit now")
			return
		}
		*/
		// 黑名单机制开启
		if Conf.Base.Blacklist == 1{
			isMelanismIp := CheckUserMelanism(ipPort)
			if isMelanismIp {
				libs.ZapLogger.Error("the Melanism ip is "+ipPort)
				return
			}
			isMelanismUid := CheckUserMelanism(mfwUid.Value)
			if isMelanismUid {
				libs.ZapLogger.Error("the Melanism uid is "+mfwUid.Value)
				return
			}
		}
		// 写入配置
		cl = NewClient(server.BroadcastSize,server.SignalSize)
		cl.conn = conn
		// 将连接保存到本机字典
		cl.Uid = mfwUid.Value
		cl.Time = libs.UnixTime(time.Now())
		cl.Uuid = strconv.FormatInt(cl.Time, 10) + cl.Uid

		store(cl.Uuid, cl)
		libs.ZapLogger.Info("connNum is "+ strconv.Itoa(connNumber))
		//hash insert to redis
		SaveUserInfo(cl.Uid, cl.Uuid,platform,role)
	} else {
		libs.ZapLogger.Info("have no mfw_uid,so quit now")
		return
	}

	go cl.HeartBeat()
	go server.writePump(cl)
	go server.readPump(cl)
}

func (s *Server) readPump(cl *Client) {
	defer func() {
		libs.ZapLogger.Info("quit readPump")
		cl.Release()
	}()

	cl.conn.SetReadLimit(s.MaxMessageSize)
	//cl.conn.SetReadDeadline(time.Now().Add(s.Options.PongWait))
	/*
		cl.conn.SetPongHandler(func(string) error {
			cl.conn.SetReadDeadline(time.Now().Add(s.Options.PongWait))
			return nil
		})
	*/

	for {
		_, message, err := cl.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				libs.ZapLogger.Error("readPump ReadMessage err:"+err.Error())
				return
			}
		}
		if message == nil {
			return
		}
		var (
			msg *libs.Message
		)
		if err := jsoniter.Unmarshal([]byte(message), &msg); err != nil {
			libs.ZapLogger.Error("message struct err: It is not a json string")
			return
		}
		switch msg.Action {
		case libs.OP_DISCONNECT:
			libs.ZapLogger.Info("initiative disconnect")
			cl.Release()
		case libs.OP_RECEIVE_ACK:
			libs.ZapLogger.Info("got the send msg ack")
			cl.receiveAck <- 1
			cl.receiveHeart <- 1
			/*
		case libs.OP_HEART_BEAT_ACK:
			cl.receiveHeart <- 1
			libs.ZapLogger.Info("got the client heart ack")
			*/
		case libs.OP_CLIENT_PING:
			tob := &libs.Proto{Ver: 1, Operation: libs.OP_CLIENT_PING, Body: string(message)}
			cl.broadcast <- tob
			cl.receiveHeart <- 1
			libs.ZapLogger.Info("got the client ping frame ,notice the heartbeat.")
			libs.ZapLogger.Info("connNum is "+ strconv.Itoa(connNumber))
		default:
			tob := &libs.Proto{Ver: 1, Operation: libs.OP_SINGLE_SEND, Body: string(message)}
			cl.broadcast <- tob
			cl.receiveHeart <- 1
			libs.ZapLogger.Info("got the other ping frame ,notice the heartbeat.")
		}
	}
}

func (s *Server) writePump(cl *Client) {
	defer func() {
		if err := recover(); err != nil {
			libs.ZapLogger.Error(err.(string))
			libs.ZapLogger.Info("quit writePump")
			cl.Release()
		}
	}()
	for {
		select {
		case message, ok := <-cl.broadcast:
			cl.conn.SetWriteDeadline(time.Now().Add(s.WriteWait))
			if !ok {
				// The hub closed the channel.
				libs.ZapLogger.Warn("SetWriteDeadline not ok")
				cl.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := cl.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				libs.ZapLogger.Warn("NextWriter err is "+err.Error())
				return
			}
			libs.ZapLogger.Info("Operation is "+string(message.Operation))
			switch message.Operation {

			case libs.OP_SINGLE_SEND:

				w.Write([]byte(message.Body))

				if err := w.Close(); err != nil {
					return
				}

				go cl.Retransmission(message.Body)
			case libs.OP_CLIENT_PING:

				msgPong := `{"action":8}`
				w.Write([]byte(msgPong))
				if err := w.Close(); err != nil {
					return
				}
			}
			libs.ZapLogger.Info("send msg is "+string(message.Body))
			cl.startHeart <- 1
			/*
		case <-time.After(time.Second * time.Duration(15)):
			//cl.startHeart <- 1
			*/
		case <-cl.writeClose:
			return
		}
	}
}
