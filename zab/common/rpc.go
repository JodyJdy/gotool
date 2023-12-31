package common

import "net/rpc"

type ZabRpc interface {
	// VoteNotification 投票信息
	VoteNotification(request VoteNotification, reply *struct{}) error
	// Ping Leader 向 其他节点 发送ping
	Ping(msg PingMsg, reply *PingResponse) error

	// Recovery 故障恢复
	Recovery(log AppendLog, result *AppendLogResult) error

	// BroadCast 广播阶段
	BroadCast(log AppendLog, result *AppendLogResult) error

	AddLog(logCommand LogCommand, reply *struct{}) error
}

type ZabRpcClient struct {
	*rpc.Client
	// 被调用的方法路径
	Target string
}

func (c ZabRpcClient) VoteNotification(request VoteNotification, reply *struct{}) error {
	return c.Client.Call(c.Target+".VoteNotification", request, reply)
}

func (c ZabRpcClient) Ping(msg PingMsg, reply *PingResponse) error {
	return c.Client.Call(c.Target+".Ping", msg, reply)
}
func (c ZabRpcClient) Recovery(log AppendLog, result *AppendLogResult) error {
	return c.Client.Call(c.Target+".Recovery", log, result)
}
func (c ZabRpcClient) BroadCast(log AppendLog, result *AppendLogResult) error {
	return c.Client.Call(c.Target+".BroadCast", log, result)
}

func (c ZabRpcClient) AddLog(logCommand LogCommand, reply *struct{}) error {
	return c.Client.Call(c.Target+".AddLog", logCommand, reply)

}
