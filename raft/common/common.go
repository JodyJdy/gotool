package common

import (
	"sync"
	"sync/atomic"
)

// NodeState 枚举 节点状态
type NodeState int

// LEADER 领导者
var LEADER NodeState = 1

// CANDIDATE 候选者
var CANDIDATE NodeState = 2

// FOLLOWER 追随者
var FOLLOWER NodeState = 3

type LogEntry struct {
	// 日志索引
	LogIndex int
	// 所在的任期
	LogTerm    int
	LogCommand interface{}
}
type Raft struct {
	Id int
	// 节点状态
	State NodeState

	// 1 所有服务器上面的持久性状态
	// 当前任期
	CurrentTerm int
	// 将票投给候选者id :VoteFor
	VotedFor int
	// 日志条目
	Logs []LogEntry

	// 2  所有服务器上面的易失性状态
	// 已知已提交的最高的日志条目的索引（初始值为0，单调递增）
	CommitIndex int
	// 已知已经被应用到状态机的最高的日志条目的索引（初始值为0，单调递增）
	LastApplied int

	// 3 领导人服务器 上的易失性状态（选举后已经重新初始化）
	// 对于每一台服务器，发送到该服务器的下一个日志条目的索引（初始值为领导人最后的日志条目的索引+1）
	NextIndex []int
	// 对于每一台服务器，已知的已经复制到该服务器的最高日志条目的索引（初始值为0，单调递增）
	MatchIndex []int

	// 表明当前所处角色的行为
	Action func(r *Raft)

	// 其他节点 节点下标->Client的映射
	Peers sync.Map
	// 监听地址
	Address string
	// 其他节点地址
	OtherNodeAddress []string
	// 投票数
	VoteCount atomic.Int32

	// 更改/获取 Raft状态的锁
	Lock sync.Mutex
}

// AppendEntriesRequest 追加条目Rpc调用请求体
// 由领导人调用，用于日志条目的复制，同时也被当做心跳使用
type AppendEntriesRequest struct {
	// 领导人的任期
	Term int
	// 领导人ID; 跟随者收到请求后，可以使用领导人id将请求重定向到领导节点
	LeaderId int
	//紧邻新日志条目之前的那个日志条目的索引
	PrevLogIndex int
	//紧邻新日志条目之前的那个日志条目的任期
	PrevLogTerm int
	// 需要保存的 日志条目（被当做心跳使用时，则日志条目内容为空；为了提高效率可能一次性发送多个）
	Entries []LogEntry
	//领导人的已知已提交的最高的日志条目的索引
	LeaderCommit int
}

// AppendEntriesResponse 追加条目Rpc响应参数
type AppendEntriesResponse struct {
	// 当前任期，对于领导人而言 它会更新自己的任期
	Term int
	//如果跟随者所含有的条目和 prevLogIndex 以及 prevLogTerm 匹配上了，则为 true
	Success bool
}

// RequestVoteRequest 请求投票RPC请求体
type RequestVoteRequest struct {
	//候选人的任期号
	Term int
	//请求选票的候选人的 ID
	CandidateId int
	// 候选人的最后日志条目的索引值
	LastLogIndex int
	// 候选人最后日志条目的任期号
	LastLogTerm int
}

// RequestVoteResponse 请求投票RPC响应
type RequestVoteResponse struct {
	//当前任期号，以便于候选人去更新自己的任期号
	Term int
	//候选人赢得了此张选票时为真
	VoteGranted bool
}
