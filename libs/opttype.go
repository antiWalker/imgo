package libs

const (
	OP_SINGLE_SEND    = int32(2) // 私聊
	OP_DISCONNECT     = int32(3) // 断开连接
	/*
	OP_HEART_BEAT     = int32(5) // 服务器端发送心跳
	OP_HEART_BEAT_ACK = int32(6) // 收到客户端的心跳应答
	*/
	OP_CLIENT_PING    = int32(7) // 客户端发送心跳ping
	OP_SERVER_PONG    = int32(8) // 收到服务端的心跳应答pong
	OP_RECEIVE_ACK    = int32(9) // 消息投递应答
)
