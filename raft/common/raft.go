package common

import (
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

func (rf *Raft) getLastIndex() int {
	return rf.Logs[len(rf.Logs)-1].LogIndex
}
func (rf *Raft) getLastTerm() int {
	return rf.Logs[len(rf.Logs)-1].LogTerm
}

// RequestVote 接收到RequestVote请求
func (rf *Raft) RequestVote(request RequestVoteRequest, response *RequestVoteResponse) error {

	rf.Lock.Lock()
	defer rf.Lock.Unlock()
	// 1. 如果term < currentTerm返回 false
	response.VoteGranted = false
	if request.Term < rf.CurrentTerm {
		response.Term = rf.CurrentTerm
		return nil
	}

	// 如果请求方 的任期 比自己大， 将自己切换为 Follower
	if request.Term > rf.CurrentTerm {
		fmt.Println("切换为Follower  RequestVote")
		rf.CurrentTerm = request.Term
		rf.SetAction(Follower, FOLLOWER)
		rf.VotedFor = -1
	}

	// 判断是否要给 请求方投票，候选人的日志至少和自己一样新
	response.Term = rf.CurrentTerm
	// 通过任期和 日志索引来判断候选人日志
	term := rf.getLastTerm()
	index := rf.getLastIndex()
	uptoDate := false

	if request.LastLogTerm > term || (request.LastLogTerm == term && request.LastLogIndex >= index) {
		uptoDate = true
	}

	//2. 如果 votedFor 为空或者为 candidateId，并且候选人的日志至少和自己一样新，那么就投票给他
	if (rf.VotedFor == -1 || rf.VotedFor == request.CandidateId) && uptoDate {
		receivedVote.Store(true)
		fmt.Println("切换为Follower  RequestVote")
		rf.SetAction(Follower, FOLLOWER)
		// 投票给请求方
		response.VoteGranted = true
		rf.VotedFor = request.CandidateId
	}
	return nil
}

// AppendEntries 接收到AppendEntries请
func (rf *Raft) AppendEntries(request AppendEntriesRequest, response *AppendEntriesResponse) error {
	response.Success = false
	receivedHeatBeat.Store(true)
	// 1. 返回假 如果领导人的任期小于接收者的当前任期
	if request.Term < rf.CurrentTerm {
		response.Term = rf.CurrentTerm
		response.NextIndex = rf.getLastIndex() + 1
		return nil
	}
	// 所有 服务的要求； 请求中的 任期比自己大，要更新当前的任期
	if request.Term > rf.CurrentTerm {
		rf.CurrentTerm = request.Term
		rf.Lock.Lock()
		fmt.Println("切换为Follower  AppendEntries")
		rf.SetAction(Follower, FOLLOWER)
		rf.VotedFor = -1
		rf.Lock.Unlock()
	}
	response.Term = request.Term

	// 如果已经超过了当前节点最大的索引，需要调整，这样下次，服务端就会发送正常的数据
	if request.PrevLogIndex > rf.getLastIndex() {
		response.NextIndex = rf.getLastIndex() + 1
		fmt.Printf("直接结束: %d %d \n", request.PrevLogIndex, response.NextIndex)
		return nil
	}
	// 2. 返回假 如果接收者日志中没有包含这样一个条目 即该条目的任期在 prevLogIndex 上能和 prevLogTerm 匹配上
	//（译者注：在接收者日志中 如果能找到一个和 prevLogIndex 以及 prevLogTerm 一样的索引和
	//任期的日志条目 则继续执行下面的步骤 否则返回假）
	baseIndex := rf.Logs[0].LogIndex
	if request.PrevLogIndex > baseIndex {
		term := rf.Logs[request.PrevLogIndex-baseIndex].LogTerm
		//从节点的 日志索引与任期 与 主节点的不一致
		if request.PrevLogTerm != term {
			// 对 NextIndex值进行调整，向前移动一个索引
			response.NextIndex = request.PrevLogIndex
			return nil
		}
	}

	//3. 如果一个已经存在的条目和新条目（译者注：即刚刚接收到的日志条目）发生了冲突（因为索引相同，任期不同），那
	//么就删除这个已经存在的条目以及它之后的所有条目
	// 4. 追加日志中尚未存在的任何新条目
	// 这里直接将PrevLogIndex后面的去掉，然后替换为新的
	rf.Logs = rf.Logs[:request.PrevLogIndex+1-baseIndex]
	if len(request.Entries) > 0 {
		fmt.Println("追加日志：")
		fmt.Println(request.Entries)
	}
	rf.Logs = append(rf.Logs, request.Entries...)
	response.Success = true
	response.NextIndex = rf.getLastIndex() + 1

	//5.  如果领导人的已知已提交的最高日志条目的索引大于接收者的已知已提交最高日志条目的索引（leaderCommit > commitIndex），
	//则把接收者的已知已经提交的最高的日志条目的索引commitIndex 重置为 领导人的已知已经提交的最高的日志条目的索引
	//leaderCommit 或者是 上一个新条目的索引 取两者的最小值
	if request.LeaderCommit > rf.CommitIndex {
		rf.CommitIndex = Min(rf.getLastIndex(), request.LeaderCommit)
	}
	return nil
}

// LeaderFirst 第一次选取为Leader时执行的行为
func LeaderFirst(r *Raft) {
	fmt.Println("leader first")
	r.Lock.Lock()
	r.NextIndex = make([]int, len(r.OtherNodeAddress))
	r.MatchIndex = make([]int, len(r.OtherNodeAddress))
	// 初始化 NextIndex,MatchIndex信息
	for i := range r.OtherNodeAddress {
		r.NextIndex[i] = r.getLastIndex() + 1
		r.MatchIndex[i] = 0
	}
	r.Action = Leader
	r.Lock.Unlock()
	Leader(r)
}

// Leader 职责
func Leader(rf *Raft) {
	//fmt.Println("leader ")
	// 更新 commitIndex, 如果提交到了大多数服务器，就认为可以提交了
	rf.updateCommitIndex()

	// 引入 baseIndex，是因为Logs[0]之前的日志可能存入了快照；此时的baseIndex不为0
	baseIndex := rf.Logs[0].LogIndex
	for i := 0; i < len(rf.OtherNodeAddress); i++ {
		if rf.NextIndex[i] > baseIndex {
			var request AppendEntriesRequest
			request.Term = rf.CurrentTerm
			request.LeaderId = rf.Id
			request.PrevLogIndex = rf.NextIndex[i] - 1
			// PrevLogIndex-baseIndex 表示的是PrevLogIndex对应的LogEntry在 rf.Logs中的下标
			request.PrevLogTerm = rf.Logs[request.PrevLogIndex-baseIndex].LogTerm
			// 将PrevLogIndex 后面的日志 全量发送过去
			request.Entries = make([]LogEntry, len(rf.Logs[request.PrevLogIndex+1-baseIndex:]))
			copy(request.Entries, rf.Logs[request.PrevLogIndex+1-baseIndex:])
			request.LeaderCommit = rf.CommitIndex
			go func(i int, req AppendEntriesRequest) {
				rf.sendAppendEntries(i, req, new(AppendEntriesResponse))
			}(i, request)
		}

	}
	time.Sleep(50 * time.Millisecond)
}
func (r *Raft) updateCommitIndex() {
	MaxCommitIndex := r.CommitIndex
	last := r.getLastIndex()
	baseIndex := r.Logs[0].LogIndex
	for i := r.CommitIndex + 1; i <= last; i++ {
		// 判断这个 索引 有没有在大多数节点上提交
		num := 0
		for _, index := range r.MatchIndex {
			// MatchIndex 表示 从节点已经接受的日志索引； 还需要是当前任期内的
			if index >= i && r.Logs[i-baseIndex].LogTerm == r.CurrentTerm {
				num++
			}
		}
		//大多数节点已经提交了
		if num >= len(r.OtherNodeAddress)/2 {
			MaxCommitIndex = i
		}
	}
	r.CommitIndex = MaxCommitIndex
}

// 是否接收到leader的心跳
var receivedHeatBeat atomic.Bool

// 是否接收到投票请求
var receivedVote atomic.Bool

// Follower 追随者职责
func Follower(r *Raft) {
	fmt.Println("Follower")
	receivedHeatBeat.Store(false)
	receivedVote.Store(false)
	time.Sleep(time.Duration(rand.Int63()%333+550) * time.Millisecond)
	// 表示没有接收到请求， 转换为 Candidate
	if !receivedHeatBeat.Load() && !receivedVote.Load() {
		r.Lock.Lock()
		r.SetAction(Candidate, CANDIDATE)
		r.Lock.Unlock()
	}
}

// Candidate 候选者职责
func Candidate(r *Raft) {
	fmt.Println("Candiate")
	receivedHeatBeat.Store(false)
	r.Lock.Lock()
	// 开始选举时，任期加1
	r.CurrentTerm++
	//投票给自己
	r.VotedFor = r.Id
	// 初始化投票数为1
	r.VoteCount.Store(1)
	// 发送投票请求，收集此时的任期和日志索引信息
	var args RequestVoteRequest
	args.Term = r.CurrentTerm
	args.CandidateId = r.Id
	args.LastLogTerm = r.getLastTerm()
	args.LastLogIndex = r.getLastIndex()
	r.Lock.Unlock()

	go func() {
		for i := 0; i < len(r.OtherNodeAddress); i++ {
			if r.State == CANDIDATE {
				go func(i int) {
					var reply RequestVoteResponse
					r.sendRequestVote(i, args, &reply)
				}(i)
			}
		}
	}()
	time.Sleep(time.Duration(rand.Int63()%300+510) * time.Millisecond)
	if r.State == LEADER {
		return
	}
	// 收到心跳请求，切换为Follower
	if receivedHeatBeat.Load() && r.State != FOLLOWER {
		fmt.Println("切换为Follower Candidate  ")
		r.SetAction(Follower, FOLLOWER)
	}
}
func getRpcClient(rf *Raft, server int) *RaftRpcClient {
	client, ok := rf.Peers.Load(server)
	if ok {
		return reflect.ValueOf(client).Interface().(*RaftRpcClient)
	}
	return nil
}
func setRpcClient(rf *Raft, server int, client *RaftRpcClient) {
	rf.Peers.Store(server, client)
}
func (rf *Raft) sendRequestVote(server int, request RequestVoteRequest, reply *RequestVoteResponse) bool {
	rf.Peers.Load(server)
	client := getRpcClient(rf, server)
	// 尝试重连
	if client == nil {
		rf.ResetConn(server)
	}
	client = getRpcClient(rf, server)
	// 重连失败，结束
	if client == nil {
		return false
	}
	err := client.RequestVote(request, reply)
	if err != nil {
		//下次重新建立连接
		setRpcClient(rf, server, nil)
		return false
	}
	rf.Lock.Lock()
	defer rf.Lock.Unlock()
	// 如果响应中的term 大于当前term， 转换为follower
	if reply.Term > rf.CurrentTerm {
		rf.CurrentTerm = reply.Term
		fmt.Println("切换为Follower sendsendRequestVote")
		rf.SetAction(Follower, FOLLOWER)
		rf.VotedFor = -1
	}
	if reply.VoteGranted {
		rf.VoteCount.Add(1)
		x := len(rf.OtherNodeAddress)
		if rf.VoteCount.Load() > int32(x) {

		}
		if rf.State == CANDIDATE && rf.VoteCount.Load() > int32(len(rf.OtherNodeAddress)/2) {
			rf.SetAction(LeaderFirst, LEADER)
		}
	}
	return true
}
func (rf *Raft) sendAppendEntries(server int, request AppendEntriesRequest, reply *AppendEntriesResponse) {
	client := getRpcClient(rf, server)
	// 尝试重连
	if client == nil {
		rf.ResetConn(server)
	}
	client = getRpcClient(rf, server)
	// 重连失败，结束
	if client == nil {
		return
	}
	err := client.AppendEntries(request, reply)
	if err != nil {
		//下次尝试重新建立连接
		setRpcClient(rf, server, nil)
	}

	// 调整为 Follower
	if reply.Term > rf.CurrentTerm {
		fmt.Println("切换为 Follower sendAppendEntries")
		fmt.Println(reply.Term)
		fmt.Println(rf.CurrentTerm)
		rf.Lock.Lock()
		rf.CurrentTerm = reply.Term
		rf.VotedFor = -1
		rf.SetAction(Follower, FOLLOWER)
		rf.Lock.Unlock()
		return
	}
	// 日志同步成功，由于将NextIndex之后的日志全量拷贝了过去
	if reply.Success {
		if len(request.Entries) > 0 {
			lastEntryIndex := request.Entries[len(request.Entries)-1].LogIndex
			rf.NextIndex[server] = lastEntryIndex + 1
			rf.MatchIndex[server] = lastEntryIndex
		}
	} else {
		fmt.Printf("同步失败 %d  %d", rf.NextIndex[server], reply.NextIndex)
		// 更新失败时 reply中包含了NextIndex,进行调整
		rf.NextIndex[server] = reply.NextIndex
	}
}

// StartListen 开启监听
func (rf *Raft) StartListen() {
	// 注册服务
	rpc.RegisterName("Raft", rf)
	listener, _ := net.Listen("tcp", rf.Address)
	go func() {
		for {
			conn, _ := listener.Accept()
			go rpc.ServeConn(conn)
		}
	}()
}

func (rf *Raft) InitRpcClient() {
	rf.Peers = sync.Map{}
	for i, v := range rf.OtherNodeAddress {
		conn, err := net.Dial("tcp", v)
		fmt.Println(err)
		if err == nil {
			cli := rpc.NewClient(conn)
			setRpcClient(rf, i, &RaftRpcClient{
				Client: cli,
				Target: "Raft",
			})
		}
	}
}
func (rf *Raft) CheckAndResetConn() {
	for i, v := range rf.OtherNodeAddress {
		if getRpcClient(rf, i) == nil {
			conn, err := net.Dial("tcp", v)
			if err == nil {
				cli := rpc.NewClient(conn)
				setRpcClient(rf, i, &RaftRpcClient{
					Client: cli,
					Target: "Raft",
				})
			}
		}
	}
}

// ResetConn 重置某个连接
func (rf *Raft) ResetConn(peer int) {
	conn, err := net.Dial("tcp", rf.OtherNodeAddress[peer])
	if err == nil {
		cli := rpc.NewClient(conn)
		setRpcClient(rf, peer, &RaftRpcClient{
			Client: cli,
			Target: "Raft",
		})
	}
}
func NewRaft() *Raft {
	var raft Raft
	raft.State = FOLLOWER
	raft.CurrentTerm = 0
	raft.VotedFor = -1
	raft.Logs = []LogEntry{
		{
			LogIndex: 0,
			LogTerm:  0,
		},
	}
	raft.CommitIndex = 0
	raft.LastApplied = 0
	raft.Action = Follower
	return &raft

}

func (rf *Raft) SetAction(f func(r *Raft), state NodeState) {
	rf.Action = f
	rf.State = state
}

// StartMainLoop 执行 raft主循环
func (rf *Raft) StartMainLoop() {
	for {
		rf.Action(rf)
	}
}
