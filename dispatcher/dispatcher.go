package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/panjf2000/ants"
	"go.uber.org/zap"
	"imgo/libs"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

var (
	GoPool *ants.PoolWithFunc
)

type Request struct {
	ToUid       string
	Content     string
	Connections map[string]string
	Result      chan string
}

func main() {
	flag.Parse()

	libs.InitLogger("../logs/dispatcher.log", "dispatcher")
	defer libs.ZapLogger.Sync()

	if err := InitConfig(); err != nil {
		errstr := "Fatal error config file: " + err.Error()
		libs.ZapLogger.Error(errstr)
	}
	libs.ZapLogger.Sugar().Infof("conf :%v", Conf)

	if Conf.Base.UsePool == 1 {
		GoPool, _ = ants.NewPoolWithFunc(Conf.Base.PoolSize, handleRequest)
		defer GoPool.Release()
	}

	// 设置cpu 核数
	runtime.GOMAXPROCS(Conf.Base.MaxProc)

	// 初始化redis
	if err := InitRedis(); err != nil {
		libs.ZapLogger.Panic("InitRedis() failed", zap.String("err", err.Error()))
	}
	// 初始化logic对应的 多台 logic server
	if err := InitRpcConnect(); err != nil {
		libs.ZapLogger.Panic("InitRpcConnect() failed", zap.String("err", err.Error()))
	}
	//server deal data
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	s := &http.Server{
		Addr:           Conf.Base.HostAddr,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	router.POST("/sendmsg", func2)
	s.ListenAndServe()

	//下面是个通过rpcx把数据打到logic层的例子，我在logic层起了两个rpc-server。ip相同，端口不同。
	//router层拿到php传过来的uid后，查询这个uid在哪个logic上。分别使用不同的rpcClient来发送数据。

	//select {}
}

func handleRequest(payload interface{}) {
	request, ok := payload.(*Request)
	if !ok {
		libs.ZapLogger.Error("payload.(*Request) param not exist")
		return
	}

	for uuid, v := range request.Connections {
		if len(uuid) == 0 || len(v) == 0 {
			libs.ZapLogger.Error("len(uuid) == 0 || len(v) == 0")
			continue
		}
		serverid, err := strconv.Atoi(v[:1])
		if err != nil {
			libs.ZapLogger.Error("strconv.Atoi(v) err")
			continue
		}
		RpcClient, ok := RpcClientList[int16(serverid)]
		if !ok {
			libs.ZapLogger.Error("RpcClientList[int16(serverid)] !ok")
			continue
		}

		PushSingleToWorker(RpcClient, uuid, request.Content)
	}
}

func func2(c *gin.Context) {
	// 回复一个200OK,在client的http-get的resp的body中获取数据
	touid := c.PostForm("touid")
	content := c.PostForm("content")
	if len(touid) == 0 {
		errstr := "Param len(touid) == 0"
		libs.ZapLogger.Error(errstr)
		c.JSON(http.StatusOK, gin.H{
			"handled":  0,
			"errorstr": errstr,
		})
		return
	}
	if len(content) == 0 {
		errstr := "Param len(content) == 0"
		libs.ZapLogger.Error(errstr)
		c.JSON(http.StatusOK, gin.H{
			"handled":  0,
			"errorstr": errstr,
		})
		return
	}
	connections, _ := GetUserPlace(touid)
	if len(connections) == 0 {
		errstr := "touid: " + touid + " no websocket connection"
		libs.ZapLogger.Error(errstr)
		c.JSON(http.StatusOK, gin.H{
			"handled":  0,
			"errorstr": errstr,
		})
		return
	}
	request := &Request{ToUid: touid, Content: content, Connections: connections}
	if Conf.Base.UsePool == 1 && GoPool != nil {
		if err := GoPool.Invoke(request); err != nil {
			libs.ZapLogger.Error(err.Error())
			c.JSON(http.StatusOK, gin.H{
				"handled":  0,
				"errorstr": err.Error(),
			})
			return
		}
	} else {
		handleRequest(request)
	}

	c.JSON(http.StatusOK, gin.H{
		"handled":  1,
		"errorstr": "",
	})
}

//testEventUpload 测试日志上报
//暂时是上报到mobile_event，后期可以优化
func testEventUpload() {
	basic := libs.EventBasic{
		AppCode:       "win.mfw.flclient",
		AppVer:        "3.0.6",
		DeviceType:    "win",
		SysVer:        "10.0",
		HardwareModel: "64",
		EventCode:     "launch_begin",
		EventTime:     time.Now().Unix(),
		EventGuid:     "1D29A5F3-C556-47BA-A4753B50CA03B9E1",
		Uid:           "",
		OpenUdid:      "763d5fbc-4113-4611-a10f-316c3f50c3ef",
		LaunchGuid:    "E480ADC9-8CF8-4023-B468EC805AA27742",
	}
	attr := make(map[string]string)
	attr["bg"] = "aaa"
	attr["dsdsd"] = "bbb"

	bret := libs.EventUpload(basic, attr)
	if !bret {
		return
	}
}
