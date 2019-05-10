package libs

const (
	SuccessReply    = 0
	SuccessReplyMsg = "success"
)

type NoReply struct {
}

type RpcSuccessReply struct {
	Code int    `int:"code"`
	Msg  string `json:"msg"`
}
