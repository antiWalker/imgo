module dispatcher

go 1.12

replace (
	imgo/libs v0.0.0 => ../libs
	labix.org/v2/mgo v0.0.0-20140701140051-000000000287 => gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce
)

require (
	github.com/gin-contrib/sse v0.0.0-20190301062529-5545eab6dad3 // indirect
	github.com/gin-gonic/gin v1.3.0
	github.com/go-redis/redis v6.15.2+incompatible
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0 // indirect
	github.com/json-iterator/go v1.1.6
	github.com/panjf2000/ants v4.0.2+incompatible
	github.com/sirupsen/logrus v1.4.1 // indirect
	github.com/smallnest/rpcx v0.0.0-20190314105900-7f0308df0c1f
	github.com/spf13/viper v1.3.2
	github.com/ugorji/go v1.1.2 // indirect
	go.uber.org/zap v1.9.1
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v8 v8.18.2 // indirect
	imgo/libs v0.0.0
)
