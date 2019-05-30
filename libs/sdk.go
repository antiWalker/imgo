// Configuration Centers - Go language version if SDK
package libs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	ENV        = "K8S_CLUSTER_NAME"
	REMOTEADDR = "skipper.console.ab"

	APPNAMESPACE  = "APP_NAMESPACE"
	APPNAME       = "APP_NAME"
	COMMONAPPNAME = "mcommon.unification"
	REDISPREFIX   = "MAResource_REDIS_VIP_"
	DBPREFIX      = "MAResource_"
	PIKAPREFIX    = "MAResource_PIKA_VIP_"
)

var (
	url       = "http://%s:8987/getdata/%s/%s"
	configDir = "/mfw_data/operation/config/"
)

type SkipperResponse struct {
	Code     int         `json:"code"`
	Data     interface{} `json:"data"`
	ErrorMsg string      `json:"error_msg"`
}

type RedisResponse struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// Get the value form the configuration center through the application name and
// key, allowing setting the default value, Return the default value when the value
// is not obtained from the configuration center and local file, and return the value
// of the key in the configuration center instead.
func GetFloat64(key string, application ...string) (float64, error) {
	var app string
	if len(application) == 0 {
		if getAppFromEnv() == "" {
			return 0, errors.New("get application from env with error")
		}
		app = getAppFromEnv()
	} else {
		app = application[0]
	}
	retClient, err, ok := getFromClient(app, key)
	if ok {
		log.Println(err)
		retFile, err := getFromFile(app, key)
		if err != nil {
			return 0, err
		}
		return assertFloat64Type(retFile), nil
	} else {
		return assertFloat64Type(retClient), err
	}
}

func GetString(key string, application ...string) (string, error) {
	var app string
	if len(application) == 0 {
		if getAppFromEnv() == "" {
			return "", errors.New("get application from env with error")
		}
		app = getAppFromEnv()
	} else {
		app = application[0]
	}
	retClient, err, ok := getFromClient(app, key)
	if ok {
		log.Println(err)
		retFile, err := getFromFile(app, key)
		if err != nil {
			return "", err
		}
		return assertStringType(retFile), nil
	} else {
		return assertStringType(retClient), err
	}
}

func GetDbConf(dbKey string, application ...string) (string, error) {
	var app string
	if len(application) == 0 {
		if getAppFromEnv() == "" {
			return "", errors.New("get application from env with error")
		}
		app = getAppFromEnv()
	} else {
		app = application[0]
	}
	dbKey = DBPREFIX + dbKey
	key := os.Getenv(dbKey)
	if key != "" {
		dbKey = key
	}
	return GetString(dbKey, app)
}

func GetRedisConf(redisGroupName string, port ...int) (string, error) {
	var portNum int
	if len(port) == 0 {
		portNum = 6379
	} else {
		portNum = port[0]
	}
	redisGroupName = REDISPREFIX + redisGroupName
	ret, err := GetString(redisGroupName, COMMONAPPNAME)
	if err != nil {
		return "", err
	}
	var redisRet RedisResponse
	err = json.Unmarshal([]byte(ret), &redisRet)
	if err != nil {
		return "", err
	}
	if redisRet.Port == 0 {
		redisRet.Port = portNum
	}
	retStr, _ := json.Marshal(redisRet)
	return string(retStr), nil
}

func GetPikaConf(redisGroupName string, port ...int) (string, error) {
	var portNum int
	if len(port) == 0 {
		portNum = 9221
	} else {
		portNum = port[0]
	}
	redisGroupName = PIKAPREFIX + redisGroupName
	ret, err := GetString(redisGroupName, COMMONAPPNAME)
	if err != nil {
		return "", err
	}
	var pikaRet RedisResponse
	err = json.Unmarshal([]byte(ret), &pikaRet)
	if err != nil {
		return "", err
	}
	if pikaRet.Port == 0 {
		pikaRet.Port = portNum
	}
	retStr, _ := json.Marshal(pikaRet)
	return string(retStr), nil
}

func getAppFromEnv() string {
	appNamespace := os.Getenv(APPNAMESPACE)
	appName := os.Getenv(APPNAME)
	if appNamespace == "" || appName == "" {
		return ""
	}
	return appNamespace + "." + appName
}

func assertFloat64Type(v interface{}) float64 {
	switch t := v.(type) {
	default:
		return 0
	case float64:
		return t
	}
}

func assertStringType(v interface{}) string {
	switch t := v.(type) {
	default:
		return ""
	case string:
		return t
	}
}

func getFromClient(application, key string) (interface{}, error, bool) {
	var geturl string
	var clientTimeout int64
	if os.Getenv(ENV) != "" {
		geturl = fmt.Sprintf(url, "127.0.0.1", application, key)
		clientTimeout = 100
	} else {
		geturl = fmt.Sprintf(url, REMOTEADDR, application, key)
		clientTimeout = 800
	}
	clientSet := &http.Client{
		Timeout: time.Duration(time.Duration(clientTimeout) * time.Millisecond),
	}
	req, err := http.NewRequest("GET", geturl, nil)
	if err != nil {
		return "", err, true
	}
	resp, err := clientSet.Do(req)
	if err != nil {
		return "", err, true
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(fmt.Sprintf("request url %s return code is %d", geturl, resp.StatusCode)), false
	}
	result, _ := ioutil.ReadAll(resp.Body)
	ret := &SkipperResponse{}
	json.Unmarshal(result, ret)
	if ret.Code == 1 {
		return "", errors.New(ret.ErrorMsg + ". key=" + key), false
	}
	return ret.Data, nil, false
}

// Gets the configuration information from the local file, returns error if the local file does
// not exist, and returns string type configuration
func getFromFile(application, key string) (interface{}, error) {
	filePath := filepath.Join(configDir, application)
	_, err := os.Stat(filePath)
	if err != nil {
		// The reason for the error is usually that the file does not exists or
		// does not have file permissions
		return "", err
	}
	fi, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer fi.Close()
	jsonStr, err := ioutil.ReadAll(fi)
	var mapResult map[string]interface{}
	if err := json.Unmarshal(jsonStr, &mapResult); err != nil {
		return "", err
	}
	if value, ok := mapResult[key]; ok {
		return value, nil
	} else {
		return "", errors.New("key is not exist")
	}
}
