module logic

go 1.12

replace (
	imgo/libs v0.0.0 => ../libs
	labix.org/v2/mgo v0.0.0-20140701140051-000000000287 => gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce
)

require (
	github.com/go-redis/redis v6.15.2+incompatible
	github.com/gorilla/websocket v1.4.0
	github.com/json-iterator/go v1.1.6
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/smallnest/rpcx v0.0.0-20190314105900-7f0308df0c1f
	github.com/spf13/viper v1.3.2
	imgo/libs v0.0.0
)
