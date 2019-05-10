package libs

type PushMsgArg struct {
	Uuid string
	Msg  string
}

type Message struct {
	Action int32 `json:"action"`
}
