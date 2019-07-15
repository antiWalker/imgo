package main

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"imgo/benckmark/http/requester"
	"math"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var (
	Conf     *Config
	confPath string
)

func init() {
	flag.StringVar(&confPath, "d", "../../config/", " set httpbenckmark config file path")
}

type Reqbody struct {
	Body string `mapstructure:"body"`
}

type Config struct {
	Base     BaseConf  `mapstructure:"base"`
	ReqBodys []Reqbody `mapstructure:"reqbodys"`
}

// 基础的配置信息
type BaseConf struct {
	Duration    int    `mapstructure:"duration"`
	Threadnum   int    `mapstructure:"threadnum"` //
	QPS         int    `mapstructure:"qps"`       //
	URL         string `mapstructure:"url"`       //
	Method      string `mapstructure:"method"`
	ContentType string `mapstructure:"contenttype"`
	Timeout     int    `mapstructure:"timeout"` //
}

func InitConfig() (err error) {
	viper.SetConfigName("httpbenckmark")
	viper.SetConfigType("toml")
	viper.AddConfigPath(confPath)

	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	if err := viper.Unmarshal(&Conf); err != nil {
		panic(fmt.Errorf("unable to decode into struct：  %s \n", err))
	}

	return nil
}

func main() {
	InitConfig()

	req, _ := http.NewRequest(Conf.Base.Method, Conf.Base.URL, nil)
	header := make(http.Header)
	header.Set("Content-Type", Conf.Base.ContentType)
	req.Header = header
	var reqbodys []string
	for _, v := range Conf.ReqBodys {
		reqbodys = append(reqbodys, v.Body)
	}
	w := &requester.Work{
		Request:  req,
		ReqBodys: reqbodys,
		N:        math.MaxInt32,
		C:        Conf.Base.Threadnum,
		Timeout:  Conf.Base.Timeout,
		QPS:      float64(Conf.Base.QPS),
		//Output:   "csv",	//hey库暂只支持csv或者控制台打印结果。2者只能输出其一，设置该参数后仅输出到benckmark.csv
	}
	w.Init()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		w.Stop()
	}()
	if Conf.Base.Duration > 0 {
		go func() {
			var duration int64 = int64(Conf.Base.Duration) * 1e9
			time.Sleep(time.Duration(duration))
			w.Stop()
		}()
	}
	w.Run()
}
