package common

import "net/rpc"

// RaftRpc raft使用到的rpc方法
type RaftRpc interface {
	// RequestVote 请求投票
	RequestVote(request RequestVoteRequest, response *RequestVoteResponse) error
	// AppendEntries 追加条目
	AppendEntries(request AppendEntriesRequest, response *AppendEntriesResponse) error

	// InstallSnapshot 安装快照
	InstallSnapshot(request InstallSnapshotRequest, response *InstallSnapshotResponse) error

	// AppendLog 由客户端调用 向 Leader添加数据
	AppendLog(LogCommand interface{}, reply *struct{}) error
}

type RaftRpcClient struct {
	*rpc.Client
	// 被调用的方法路径
	Target string
}
type AppendLogError struct {
	Msg           error
	LeaderAddress string
}

func (c RaftRpcClient) RequestVote(request RequestVoteRequest, response *RequestVoteResponse) error {
	return c.Client.Call(c.Target+".RequestVote", request, response)
}
func (c RaftRpcClient) AppendEntries(request AppendEntriesRequest, response *AppendEntriesResponse) error {
	return c.Client.Call(c.Target+".AppendEntries", request, response)
}
func (c RaftRpcClient) AppendLog(logCommand LogCommand, reply *struct{}) error {
	return c.Client.Call(c.Target+".AppendLog", logCommand, reply)
}
func (c RaftRpcClient) InstallSnapshot(request InstallSnapshotRequest, response *InstallSnapshotResponse) error {
	return c.Client.Call(c.Target+".InstallSnapshot", request, response)
}
