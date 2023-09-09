package common

import "net/rpc"

type ZabRpc interface {
	// VoteNotification 投票信息
	VoteNotification(request VoteNotification, reply *struct{}) error
	// Ping Leader 向 其他节点 发送ping
	Ping(msg PingMsg, reply *struct{}) error
}

type ZabRpcClient struct {
	*rpc.Client
	// 被调用的方法路径
	Target string
}

func (c ZabRpcClient) VoteNotification(request VoteNotification, reply *struct{}) error {
	return c.Client.Call(c.Target+".VoteNotification", request, reply)
}

func (c ZabRpcClient) Ping(msg PingMsg, reply *struct{}) error {
	return c.Client.Call(c.Target+".Ping", msg, reply)
}
