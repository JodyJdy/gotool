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
func (rf *Raft) IsLeader() bool {
	return rf.State == LEADER
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
		rf.SetAction(Follower, FOLLOWER)
		// 投票给请求方
		response.VoteGranted = true
		rf.VotedFor = request.CandidateId
	}
	return nil
}

// AppendEntries 接收到AppendEntries请求
func (rf *Raft) AppendEntries(request AppendEntriesRequest, response *AppendEntriesResponse) error {
	receivedHeatBeat.Store(true)
	return nil
}

// SendRequestVote 发送 RequestVote请求
func (rf *Raft) SendRequestVote(request RequestVoteRequest) (*RequestVoteResponse, error) {
	result := new(RequestVoteResponse)
	client := RaftRpcClient{}
	return result, client.RequestVote(request, result)
}

// SendAppendEntries 发送 AppendEntries 请求
func (rf *Raft) SendAppendEntries(request AppendEntriesRequest) (*AppendEntriesResponse, error) {
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
	r.Lock.Lock()
	r.Action = Leader
	r.Lock.Unlock()
	Leader(r)
}

// Leader 职责
func Leader(r *Raft) {
	fmt.Println("leader ")
	var request AppendEntriesRequest
	for i := 0; i < len(r.OtherNodeAddress); i++ {
		go func(i int) {
			var reply AppendEntriesResponse
			r.sendAppendEntries(i, request, &reply)
		}(i)
	}
	time.Sleep(50 * time.Millisecond)
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
	if reply.Term > rf.CurrentTerm {
		rf.CurrentTerm = reply.Term
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
