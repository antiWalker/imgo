package main

import (
	"flag"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
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
	rand.New(rand.NewSource(10))
	for i := 0; i < *connections; i++ {
		req, err := http.NewRequest("GET", "http://"+*host, nil)
		var uid string
		uid = strconv.Itoa(rand.Int())
		//strconv.Itoa(n)
		cookie1 := &http.Cookie{
			Name:  "mfw_uid",
			Value:	uid,
		}
		cookie1 = &http.Cookie{
			Name:  "mfw_uid",
			Value:	"345678",
		}
		cookie2 := &http.Cookie{
			Name:  "f",
			Value: "1",
		}
		cookie3 := &http.Cookie{
			Name:  "r",
			Value: "1",
		}
		cookie4 := &http.Cookie{
			Name:  "p",
			Value: "1",
		}
		req.AddCookie(cookie1)
		req.AddCookie(cookie2)
		req.AddCookie(cookie3)
		req.AddCookie(cookie4)
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
