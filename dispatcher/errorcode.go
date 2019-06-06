package main

const (
	PARAM_ERROR int = -10  //请求参数有误
	REDIS_ERROR int = -100 //redis连接断开
	POOL_ERROR  int = -200 //协程池申请失败
)
