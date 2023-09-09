package common

import (
	"sync"
)

// NodeState 枚举 节点状态
type NodeState int

// LEADER 领导者
var LEADER NodeState = 1

// 选举状态
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
	Lock sync.Mutex
}

func (zab *Zab) ZxId() uint64 {
	return (uint64(zab.Epoch) << 32) | uint64(zab.Counter)
}

// IncrementCounter 计数器累加
func (zab *Zab) IncrementCounter() {
	zab.Lock.Lock()
	zab.Counter++
	zab.Lock.Unlock()
}

// IncrementEpoch 纪元累加
func (zab *Zab) IncrementEpoch() {
	zab.Lock.Lock()
	zab.Epoch++
	zab.Lock.Unlock()
}
