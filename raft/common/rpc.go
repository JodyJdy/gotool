package common

import "net/rpc"

// RaftRpc raft使用到的rpc方法
type RaftRpc interface {
	// RequestVote 请求投票
	RequestVote(request RequestVoteRequest, response *RequestVoteResponse) error
	// AppendEntries 追加条目
	AppendEntries(request AppendEntriesRequest, response *AppendEntriesResponse) error
}

type RaftRpcClient struct {
	*rpc.Client
	// 被调用的方法路径
	Target string
}

func (c RaftRpcClient) RequestVote(request RequestVoteRequest, response *RequestVoteResponse) error {
	return c.Client.Call(c.Target+".RequestVote", request, response)
}
func (c RaftRpcClient) AppendEntries(request AppendEntriesRequest, response *AppendEntriesResponse) error {
	return c.Client.Call(c.Target+".AppendEntries", request, response)
}
