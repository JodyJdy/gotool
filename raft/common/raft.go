package common

import "fmt"

// RequestVote 接收到RequeVote请求
func (r *Raft) RequestVote(request RequestVoteRequest, response *RequestVoteResponse) error {
	return nil
}

// AppendEntries 接收到AppendEntries请求
func (r *Raft) AppendEntries(request AppendEntriesRequest, response *AppendEntriesResponse) error {
	return nil
}

// SendRequestVote 发送 RequestVote请求
func (r *Raft) SendRequestVote(request RequestVoteRequest) (*RequestVoteResponse, error) {
	result := new(RequestVoteResponse)
	client := RaftRpcClient{}
	return result, client.RequestVote(request, result)
}

// SendAppendEntries 发送 AppendEntries 请求
func (r *Raft) SendAppendEntries(request AppendEntriesRequest) (*AppendEntriesResponse, error) {
	result := new(AppendEntriesResponse)
	client := RaftRpcClient{}
	return result, client.AppendEntries(request, result)
}

// LeaderFirst 第一次选取为Leader时执行的行为
func LeaderFirst(r *Raft) {
	fmt.Println("leader first")
	// 做一些初始操作
	// xxxxxx
	// 之后设置 Leader函数 为 节点的行为
	r.Action = Leader
}

// Leader 职责
func Leader(r *Raft) {
	fmt.Println("leader ")
}

// Follower 追随者职责
func Follower(r *Raft) {

}

// Candidate 候选者职责
func Candidate(r *Raft) {

}
