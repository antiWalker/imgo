package main

import (
	"flag"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"time"
)

var host = flag.String("host", "wss.mafengwo.cn", "server address")
var connections = flag.Int("conn", 1, "number of ws connections")
func main() {
	start()
}
//go run wsclient.go -conn 30
func start() {
	flag.Parse()

	u := url.URL{Scheme: "ws", Host: *host, Path: "/ws/"}
	log.Printf("connecting to %s", u.String())
	var conns []*websocket.Conn
	for i := 0; i < *connections; i++ {
		req, err := http.NewRequest("GET", "http://"+*host, nil)

		cookie1 := &http.Cookie{
			Name:  "mfw_uid",
			Value: "489898",
		}
		cookie2 := &http.Cookie{
			Name:  "f",
			Value: "1",
		}
		cookie3 := &http.Cookie{
			Name:  "r",
			Value: "1",
		}
		req.AddCookie(cookie1)
		req.AddCookie(cookie2)
		req.AddCookie(cookie3)
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), req.Header)

		if err != nil {
			log.Fatal("dial:", err)
		}
		conns = append(conns, conn)
		time.Sleep(time.Millisecond*7)
		log.Println("number is %d",i)
	}

	defer func() {
		for _, c := range conns {
			c.Close()
		}
	}()
	log.Printf("完成初始化 %d 连接", len(conns))
	// 发送心跳
	tts := time.Second
	if *connections > 100 {
		tts = time.Millisecond * 5
	}
	for {
		for i := 0; i < len(conns); i++ {
			time.Sleep(tts)
			conn := conns[i]
			//conn.Write([]byte("hello world\r\n"))
			input := `{"action":7}`
			conn.WriteMessage(websocket.TextMessage, []byte(input))
		}
	}
}
