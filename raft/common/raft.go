package common

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
