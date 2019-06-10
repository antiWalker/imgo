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
	"strings"
	"time"
)

const (
	ROLE_USER     int = 1
	ROLE_BUSINESS int = 2

	CHANNEL_WEB     int = 0
	CHANNEL_ANDROID int = 1
	CHANNEL_IOS     int = 2
)

var (
	SendMsgPool   *ants.PoolWithFunc
	BatchSendPool *ants.PoolWithFunc
)

type SendMsgParam struct {
	ToUid       string
	Content     string
	ToRole      string
	Connections map[string]string
	Result      *msgHandled
}

type BatchSendParam struct {
	ToUids  string
	Content string
	ToRole  string
	Uuid    string
}

type httpReturn struct {
	Errno int         `json:"errno"`
	Error string      `json:"error"`
	Data  interface{} `json:"data"`
}

type userStatus struct {
	AppB int `json:"app_b"`
	AppC int `json:"app_c"`
	WebB int `json:"web_b"`
	WebC int `json:"web_c"`
	PcB  int `json:"pc_b"`
	PcC  int `json:"pc_c"`
}

type msgHandled struct {
	Handled int `json:"handled"`
	APP     int `json:"app"`
	WEB     int `json:"web"`
	PC      int `json:"pc"`
}

func main() {
	flag.Parse()

	defer libs.ZapLogger.Sync()

	if err := InitConfig(); err != nil {
		errstr := "Fatal error config file: " + err.Error()
		libs.ZapLogger.Error(errstr)
	}
	libs.InitLogger(Conf.Base.Logfile, "dispatcher")
	libs.ZapLogger.Sugar().Infof("conf :%v", Conf)

	if Conf.Base.UsePool == 1 {
		SendMsgPool, _ = ants.NewPoolWithFunc(Conf.Base.PoolSize, PushMsg)
		defer SendMsgPool.Release()
	}

	BatchSendPool, _ = ants.NewPoolWithFunc(Conf.Base.PoolSize, BatchSend)
	defer BatchSendPool.Release()

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
	router.POST("/sendmsg", handleSendmsg)
	router.POST("/batchsendmsg", handleBatchSendmsg)
	router.POST("/checkstatus", handleCheckstatus)
	s.ListenAndServe()

	//下面是个通过rpcx把数据打到logic层的例子，我在logic层起了两个rpc-server。ip相同，端口不同。
	//router层拿到php传过来的uid后，查询这个uid在哪个logic上。分别使用不同的rpcClient来发送数据。
}

func BatchSend(payload interface{}) {
	request, ok := payload.(*BatchSendParam)
	if !ok {
		libs.ZapLogger.Error("payload.(*BatchSendParam) param not exist")
		return
	}

	arrayuids := strings.Split(request.ToUids, ",")
	sendnum := 0
	for _, touid := range arrayuids {
		if len(touid) == 0 {
			libs.ZapLogger.Error("len(touid) == 0")
			continue
		}

		connections, err := GetUserPlace(touid)
		if err != nil {
			errstr := "redis connection break"
			libs.ZapLogger.Error(errstr)
			continue
		}

		var handled msgHandled
		request := &SendMsgParam{ToUid: touid, Content: request.Content, ToRole: request.ToRole, Connections: connections, Result: &handled}
		PushToWorker(request)
		sendnum++
	}

	libs.ZapLogger.Info("BatchSend end uuid=" + request.Uuid, zap.Int("sendnum", sendnum))
}

func PushMsg(payload interface{}) {
	request, ok := payload.(*SendMsgParam)
	if !ok {
		libs.ZapLogger.Error("payload.(*SendMsgParam) param not exist")
		return
	}

	PushToWorker(request)
}

func PushToWorker(param *SendMsgParam) {
	for uuid, v := range param.Connections {
		if len(uuid) == 0 || len(v) == 0 {
			libs.ZapLogger.Error("len(uuid) == 0 || len(v) == 0")
			continue
		}

		info, ret := GetSessionInfo(v)
		if !ret {
			libs.ZapLogger.Error("GetSessionInfo(v) err")
			continue
		}

		torole, _ := strconv.Atoi(param.ToRole)
		if torole != info.Role {
			continue
		}
		RpcClient, ok := RpcClientList[info.ServerId]
		if !ok {
			libs.ZapLogger.Error("RpcClientList[int16(serverid)] !ok ServerId is "+string(info.ServerId))
			continue
		}

		PushSingleToWorker(RpcClient, uuid, param.Content)
		param.Result.Handled = 1
		if info.Channel == CHANNEL_WEB {
			param.Result.WEB += 1
		} else if info.Channel == CHANNEL_IOS || info.Channel == CHANNEL_ANDROID {
			param.Result.APP += 1
		}
	}
}

func handleBatchSendmsg(c *gin.Context) {
	touids := c.PostForm("touids")
	content := c.PostForm("content")
	torole := c.PostForm("torole")
	var ret httpReturn
	var handled msgHandled
	ret.Data = handled
	if len(touids) == 0 {
		errstr := "Param len(touids) == 0"
		libs.ZapLogger.Error(errstr)
		ret.Errno = PARAM_ERROR
		ret.Error = errstr
		httpRet(c, ret)
		return
	}
	if len(content) == 0 {
		errstr := "Param len(content) == 0"
		libs.ZapLogger.Error(errstr)
		ret.Errno = PARAM_ERROR
		ret.Error = errstr
		httpRet(c, ret)
		return
	}
	if len(torole) == 0 {
		errstr := "Param len(torole) == 0"
		libs.ZapLogger.Error(errstr)
		ret.Errno = PARAM_ERROR
		ret.Error = errstr
		httpRet(c, ret)
		return
	}

	//检查一下redis连接是否断开
	_, err := GetUserPlace("0")
	if err != nil {
		errstr := "redis connection break"
		libs.ZapLogger.Error(errstr)
		ret.Errno = REDIS_ERROR
		ret.Error = errstr
		httpRet(c, ret)
		return
	}

	if BatchSendPool == nil {
		errstr := "BatchSendPool == nil"
		libs.ZapLogger.Error(errstr)
		ret.Errno = POOL_ERROR
		ret.Error = errstr
		httpRet(c, ret)
		return
	}

	//异步执行批量请求，用md5生成一个uuid，用来追踪请求送达情况
	nownano := time.Now().UnixNano()
	strnano := strconv.FormatInt(nownano, 10)
	uuid := libs.Md5V(strnano + touids)
	request := &BatchSendParam{ToUids: touids, Content: content, ToRole: torole, Uuid: uuid}
	if err := BatchSendPool.Invoke(request); err != nil {
		libs.ZapLogger.Error(err.Error())
		ret.Errno = POOL_ERROR
		ret.Error = err.Error()
		httpRet(c, ret)
		return
	}

	libs.ZapLogger.Info("new BatchSendmsg ToUids=" + touids + " Content=" + content + " ToRole=" + torole + " Uuid=" + uuid)
	handled.Handled = 1
	ret.Data = handled
	httpRet(c, ret)
}

func handleSendmsg(c *gin.Context) {
	touid := c.PostForm("touid")
	content := c.PostForm("content")
	torole := c.PostForm("torole")
	var ret httpReturn
	var handled msgHandled
	ret.Data = handled
	if len(touid) == 0 {
		errstr := "Param len(touid) == 0"
		libs.ZapLogger.Error(errstr)
		ret.Errno = PARAM_ERROR
		ret.Error = errstr
		httpRet(c, ret)
		return
	}
	if len(content) == 0 {
		errstr := "Param len(content) == 0"
		libs.ZapLogger.Error(errstr)
		ret.Errno = PARAM_ERROR
		ret.Error = errstr
		httpRet(c, ret)
		return
	}
	if len(torole) == 0 {
		errstr := "Param len(torole) == 0"
		libs.ZapLogger.Error(errstr)
		ret.Errno = PARAM_ERROR
		ret.Error = errstr
		httpRet(c, ret)
		return
	}
	connections, err := GetUserPlace(touid)
	if err != nil && Conf.Base.NoDBStrategy == 0 {
		errstr := "redis connection break"
		libs.ZapLogger.Error(errstr)
		ret.Errno = REDIS_ERROR
		ret.Error = errstr
		httpRet(c, ret)
		return
	}
	if err == nil && len(connections) == 0 {
		httpRet(c, ret)
		return
	}

	request := &SendMsgParam{ToUid: touid, Content: content, ToRole: torole, Connections: connections, Result: &handled}
	if Conf.Base.UsePool == 1 && SendMsgPool != nil {
		if err := SendMsgPool.Invoke(request); err != nil {
			errstr := err.Error()
			libs.ZapLogger.Error(errstr)
			ret.Errno = POOL_ERROR
			ret.Error = errstr
			httpRet(c, ret)
			return
		}
	} else {
		PushMsg(request)
	}

	ret.Data = handled
	httpRet(c, ret)
}

func handleCheckstatus(c *gin.Context) {
	uid := c.PostForm("uid")
	var status userStatus
	var ret httpReturn
	ret.Data = status
	if len(uid) == 0 {
		errstr := "Param len(uid) == 0"
		libs.ZapLogger.Error(errstr)
		ret.Errno = PARAM_ERROR
		ret.Error = errstr
		httpRet(c, ret)
		return
	}

	connections, err := GetUserPlace(uid)
	if err != nil {
		errstr := "redis connection break"
		libs.ZapLogger.Error(errstr)
		ret.Errno = REDIS_ERROR
		ret.Error = errstr
		httpRet(c, ret)
		return
	}

	for _, v := range connections {
		info, ret := GetSessionInfo(v)
		if !ret {
			continue
		}
		if info.Role == ROLE_USER {
			if info.Channel == CHANNEL_WEB {
				status.WebC = 1
			} else {
				status.AppC = 1
			}
		} else {
			if info.Channel == CHANNEL_WEB {
				status.WebB = 1
			} else {
				status.AppB = 1
			}
		}
	}

	ret.Data = status
	httpRet(c, ret)
}

func httpRet(c *gin.Context, ret httpReturn) {
	c.JSON(http.StatusOK, ret)
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
