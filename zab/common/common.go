package common

import (
	"fmt"
	"sync"
)

// NodeState 枚举 节点状态
type NodeState int

// LEADER 领导者
var LEADER NodeState = 1

// LOOKING 选举状态
var LOOKING NodeState = 2

// FOLLOWER 追随者
var FOLLOWER NodeState = 3

type LogEntry struct {
	// 消息对应的事务id
	ZxId       uint64
	LogCommand interface{}
}
type Zab struct {
	//  serverId
	ServerId int
	//节点状态
	State NodeState
	//日志条目
	Logs []LogEntry
	//纪元号
	Epoch uint32
	//计数器
	Counter uint32

	// 表明当前所处角色的行为
	Action func(r *Zab)

	// 其他节点 节点下标->Client的映射
	Peers sync.Map
	// 监听地址
	Address string
	// 其他节点地址 serverId->地址 的映射
	OtherNodeAddress map[int]string
	// 更改/获取 Zab状态的锁
	StateLock sync.Mutex
	// Vote Lock 投票相关的锁
	VoteLock sync.Mutex
	//日志追加的锁
	LogLock sync.Mutex
	//Follower 接收id更改的锁
	ReceiveIdLock sync.Mutex

	// 投票信息 serverId 投票给 -> serverId 的映射
	VoteMap map[int]*Vote

	// 上次提交的 zxid, Leader,Follower均会使用
	LastCommitZxId uint64
	//接收到的消息的最大zxid（接收到还未提交）Follower 使用
	LastReceiveZxId uint64
	//存储所有Follower 上次接收的ZxId
	LastReceiveZxIdMap map[int]uint64
}
type Vote struct {
	//获取票的 服务器id
	VotedLeaderServerId int
	//获取票的服务器的 zxid
	VotedLeaderZxId uint64
	//投票人的状态
	State NodeState
}
type PingMsg struct {
	LeaderServerId int
	ZxId           uint64
	//leader已经提交了的zxid
	LeaderCommitZxId uint64
}

// PingResponse 返回上次接收的消息
type PingResponse struct {
	//接收到的消息的最大zxid（接收到还未提交）Follower 使用
	LastReceiveZxId uint64
	//当前的Epoch
	Epoch uint32
}

func (zab *Zab) SetState(state NodeState) {
	if state == LEADER {
		fmt.Println("LEADER")
	} else if state == LOOKING {
		fmt.Println("LOOKING")
	} else if state == FOLLOWER {
		fmt.Println("FOLLOWER")
	}
	zab.State = state
}

func (zab *Zab) ZxId() uint64 {
	return (uint64(zab.Epoch) << 32) | uint64(zab.Counter)
}

// IncrementCounter 计数器累加
func (zab *Zab) IncrementCounter() {
	zab.StateLock.Lock()
	zab.Counter++
	zab.StateLock.Unlock()
}

// IncrementEpoch 纪元累加
func (zab *Zab) IncrementEpoch() {
	zab.StateLock.Lock()
	zab.Epoch++
	zab.StateLock.Unlock()
}

// VoteNotification 投票通知
type VoteNotification struct {
	//投票人 server ID
	ServeId int
	//获取票的 服务器id
	VotedLeaderServerId int
	//获取票的服务器的 zxid
	VotedLeaderZxId uint64
	//投票人的状态
	State NodeState
}

type AppendLog struct {
	Entries []LogEntry
}
type AppendLogResult struct {
	Ack bool
}
