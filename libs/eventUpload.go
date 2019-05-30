package libs

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
	neturl "net/url"
	"strconv"
	"strings"
)

const encKey string = "6f71d5b32422ac3525a96268edca48bb"

var eventUploadUrl = "https://mapi.mafengwo.cn/mobilelog/rest/EventLog/"

type EventBasic struct {
	AppCode       string `json:"app_code"`
	AppVer        string `json:"app_ver"`
	DeviceType    string `json:"device_type"`
	SysVer        string `json:"sys_ver"`
	HardwareModel string `json:"hardware_model"`
	EventCode     string `json:"event_code"`
	EventTime     int64  `json:"event_time"`
	EventGuid     string `json:"event_guid"`
	Uid           string `json:"uid"`
	OpenUdid      string `json:"open_udid"`
	LaunchGuid    string `json:"launch_guid"`
}

type EventBody struct {
	Basic EventBasic        `json:"basic"`
	Attr  map[string]string `json:"attr"`
}

type Update struct {
	Sign string `json:"sign"`
	Data string `json:"data"`
	Ts   int64  `json:"ts"`
}

type JsonData struct {
	PostStyle  string `json:"post_style"`
	UpdateData Update `json:"update"`
}

func EventUpload(eventBasic EventBasic, eventAttr map[string]string) bool {
	var eventbody = EventBody{
		Basic: eventBasic,
		Attr:  eventAttr,
	}

	var jsoniterator = jsoniter.ConfigCompatibleWithStandardLibrary
	b, err := jsoniterator.Marshal(eventbody)
	if err != nil {
		fmt.Println("error:", err)
		return false
	}
	strdata := string(b)
	fmt.Println(strdata)

	strtime := strconv.FormatInt(eventBasic.EventTime, 10)
	var signdata = encKey + strtime + strdata
	signmd5 := md5V(signdata)
	signbase64 := base64V(signmd5)

	var update = Update{
		Sign: signbase64,
		Data: strdata,
		Ts:   eventBasic.EventTime,
	}
	var jsondata = JsonData{
		PostStyle:  "default",
		UpdateData: update,
	}
	alldata, err1 := jsoniterator.Marshal(jsondata)
	if err1 != nil {
		fmt.Println("error:", err1)
		return false
	}

	//忽略对服务器端证书的校验
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: tr}
	postdata := neturl.Values{}
	postdata.Set("app_code", eventBasic.AppCode)
	postdata.Set("app_ver", eventBasic.AppVer)
	postdata.Set("jsondata", string(alldata))
	postbody := postdata.Encode()

	req, err := http.NewRequest("POST", eventUploadUrl, strings.NewReader(postbody))
	if err != nil {
		fmt.Println("error:", err1)
		return false
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error:", err1)
		return false
	}

	fmt.Println(string(body))

	return true
}

func md5V(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func base64V(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}
