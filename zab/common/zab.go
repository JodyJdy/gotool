package common

import (
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"reflect"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// 尝试重连
func getRpcClientWithReConn(zab *Zab, serverId int) *ZabRpcClient {
	client := getRpcClient(zab, serverId)
	if client == nil {
		zab.ResetConn(serverId)
		client = getRpcClient(zab, serverId)
	}
	return client
}
func getRpcClient(zab *Zab, serverId int) *ZabRpcClient {
	client, ok := zab.Peers.Load(serverId)
	if ok {
		return reflect.ValueOf(client).Interface().(*ZabRpcClient)
	}
	return nil
}
func setRpcClient(zab *Zab, serverId int, client *ZabRpcClient) {
	zab.Peers.Store(serverId, client)
}

func NewZab(serverId int) *Zab {
	zab := new(Zab)
	zab.ServerId = serverId
	zab.SetAction(Looking, LOOKING)
	zab.VoteMap = make(map[int]*Vote)
	zab.OtherNodeAddress = make(map[int]string)
	zab.Logs = []LogEntry{}
	zab.LastReceiveZxIdMap = make(map[int]uint64)
	return zab
}

// StartListen 开启监听
func (zab *Zab) StartListen() {
	// 注册服务
	err := rpc.RegisterName("Zab", zab)
	if err != nil {
		return
	}
	listener, _ := net.Listen("tcp", zab.Address)
	go func() {
		for {
			conn, _ := listener.Accept()
			go rpc.ServeConn(conn)
		}
	}()
}

func (zab *Zab) InitRpcClient() {
	zab.Peers = sync.Map{}
	for i, v := range zab.OtherNodeAddress {
		conn, err := net.Dial("tcp", v)
		fmt.Println(err)
		if err == nil {
			cli := rpc.NewClient(conn)
			setRpcClient(zab, i, &ZabRpcClient{
				Client: cli,
				Target: "Zab",
			})
		}
	}
}

func (zab *Zab) CheckAndResetConn() {
	for i, v := range zab.OtherNodeAddress {
		if getRpcClient(zab, i) == nil {
			conn, err := net.Dial("tcp", v)
			if err == nil {
				cli := rpc.NewClient(conn)
				setRpcClient(zab, i, &ZabRpcClient{
					Client: cli,
					Target: "Zab",
				})
			}
		}
	}
}

// ResetConn 重置某个连接
func (zab *Zab) ResetConn(peer int) {
	conn, err := net.Dial("tcp", zab.OtherNodeAddress[peer])
	if err == nil {
		cli := rpc.NewClient(conn)
		setRpcClient(zab, peer, &ZabRpcClient{
			Client: cli,
			Target: "Zab",
		})
	}
}

// StartMainLoop 执行 Zab主循环
func (zab *Zab) StartMainLoop() {
	for {
		//执行主循环
		zab.Action(zab)
	}
}

func (zab *Zab) SetAction(f func(r *Zab), state NodeState) {
	zab.Action = f
	zab.State = state
}

// LeaderFirst 执行刚成为Leader时的行为
func LeaderFirst(zab *Zab) {
	fmt.Println("leader first")
	//如果有必要，需要进行崩溃恢复，先关闭广播
	couldBroadCast.Store(false)
	zab.IncrementEpoch()
	//重置计数器
	zab.Counter = 0
	zab.Action = Leader
	Leader(zab)
}

var couldBroadCast atomic.Bool

func Leader(zab *Zab) {
	fmt.Println("Leader")
	go zab.ping()
	//发送日志
	zab.sendLog()
	time.Sleep(50 * time.Millisecond)
}
func (zab *Zab) Ping(msg PingMsg, response *PingResponse) error {
	//接受到心跳
	receivedHeatBeat.Store(true)
	if zab.State == LOOKING {
		zab.SetAction(Follower, FOLLOWER)
	}
	//更新节点epoch
	zab.Epoch = GetEpoch(msg.ZxId)
	zab.Counter = GetCounter(msg.ZxId)
	//更新已经提交的事务
	response.LastReceiveZxId = zab.LastReceiveZxId
	//调整已经提交事务
	zab.LastCommitZxId = Min(zab.LastReceiveZxId, msg.LeaderCommitZxId)
	return nil
}

func (zab *Zab) ping() {
	msg := PingMsg{
		LeaderServerId:   zab.ServerId,
		ZxId:             zab.ZxId(),
		LeaderCommitZxId: zab.LastCommitZxId,
	}
	//进行广播，不考虑响应
	go func(msg PingMsg) {
		for k, _ := range zab.OtherNodeAddress {
			go func(serverId int) {
				zab.sendPing(serverId, msg)
			}(k)
		}
	}(msg)
}
func (zab *Zab) sendPing(serverId int, msg PingMsg) {
	client := getRpcClientWithReConn(zab, serverId)
	if client == nil {
		return
	}
	var response = new(PingResponse)
	err := client.Ping(msg, response)
	//网络有故障
	if err != nil {
		fmt.Println(err)
		setRpcClient(zab, serverId, nil)
	} else {
		//更新上次接收的最大zxid
		zab.ReceiveIdLock.Lock()
		zab.LastReceiveZxIdMap[serverId] = response.LastReceiveZxId
		zab.ReceiveIdLock.Unlock()
	}
}

// 发送投票广播
func (zab *Zab) sendVoteNotification(serverId int, notification VoteNotification) {
	client := getRpcClientWithReConn(zab, serverId)
	//重连失败
	if client == nil {
		return
	}
	err := client.VoteNotification(notification, new(struct{}))
	//网络有故障，重连
	if err != nil {
		setRpcClient(zab, serverId, nil)
	}
}

// 发送广播通知
func (zab *Zab) notification() {
	vote := zab.VoteMap[zab.ServerId]
	// 用于广播的内容
	notification := VoteNotification{
		ServeId:             zab.ServerId,
		VotedLeaderZxId:     vote.VotedLeaderZxId,
		VotedLeaderServerId: vote.VotedLeaderServerId,
		State:               vote.State,
	}
	go func(voteNotification VoteNotification) {
		for k, _ := range zab.OtherNodeAddress {
			go func(serverId int) {
				zab.sendVoteNotification(serverId, voteNotification)
			}(k)
		}
	}(notification)
}

// 重置投票箱
func (zab *Zab) resetVote() {
	zab.VoteMap = make(map[int]*Vote)
}

// 记录节点的选票
func (zab *Zab) recordVote(request VoteNotification) {
	zab.VoteMap[request.ServeId] = &Vote{
		VotedLeaderServerId: request.VotedLeaderServerId,
		VotedLeaderZxId:     request.VotedLeaderZxId,
		State:               request.State,
	}
}

// 判断有无节点得到半数支持，返回当前节点是否是Leader
// 只在状态检查的时候进行加锁
func (zab *Zab) checkState() (leaderId int) {
	zab.StateLock.Lock()
	defer zab.StateLock.Unlock()
	//记录每个节点的获取的投票数字
	countMap := make(map[int]int)
	for _, v := range zab.VoteMap {
		if v != nil {
			countMap[v.VotedLeaderServerId]++
		}
	}
	for k, v := range countMap {
		if v > len(zab.OtherNodeAddress)/2 {
			zab.resetVote()
			//获取选票的是自己
			if k == zab.ServerId {
				zab.SetAction(LeaderFirst, LEADER)
				leaderId = k
			} else {
				zab.SetAction(Follower, FOLLOWER)
				leaderId = k
			}
			//重置选票
			zab.resetVote()
			break
		}
	}
	return
}

// VoteNotification 接收到投票广播
func (zab *Zab) VoteNotification(request VoteNotification, reply *struct{}) error {
	// 如果不在选举状态中，不做处理
	if zab.State != LOOKING {
		return nil
	}
	// 线程安全的更改投票状态
	zab.VoteLock.Lock()
	defer zab.VoteLock.Unlock()
	epoch := GetEpoch(request.VotedLeaderZxId)
	//收到looking状态的选票
	if request.State == LOOKING {
		if epoch < zab.Epoch {
			//忽略
			return nil
		} else if epoch > zab.Epoch {
			//清空选票
			zab.VoteMap = make(map[int]*Vote)
			//调整当前epoch为更新的
			zab.Epoch = epoch
			zab.VoteMap[zab.ServerId] = &Vote{
				VotedLeaderServerId: request.VotedLeaderServerId,
				VotedLeaderZxId:     request.VotedLeaderZxId,
				State:               zab.State,
			}
			//记录那个人的选票
			zab.recordVote(request)
			//广播
			zab.notification()
		} else {
			//此时 epoch相同，判断自己的选票和当前的选票哪个最优
			//1.记录当前选票
			zab.recordVote(request)
			// 2. 判断选票和自己的选票，哪个更新，
			myVote := zab.VoteMap[zab.ServerId]
			if myVote == nil {
				return nil
			}
			if request.VotedLeaderZxId > myVote.VotedLeaderZxId ||
				(request.VotedLeaderZxId == myVote.VotedLeaderZxId && request.VotedLeaderServerId > myVote.VotedLeaderServerId) {
				//比自己新，更新自己的选票
				zab.VoteMap[zab.ServerId] = &Vote{
					VotedLeaderServerId: request.VotedLeaderServerId,
					VotedLeaderZxId:     request.VotedLeaderZxId,
					State:               zab.State,
				}
				//发送广播
				zab.notification()
			}
		}
		//检查状态
		zab.checkState()
	} else {
		//收到 Follower /Leader 状态节点的选票
		// epoch相同，记录选票，推断状态
		if epoch == zab.Epoch {
			//记录选票
			zab.recordVote(request)
			//检查状态
			zab.checkState()
		} else {
			//放在OutOfElection，推断是否可以结束选举，这里不进行处理
		}
	}
	return nil
}

func Looking(zab *Zab) {
	fmt.Printf("LOOKING Epoch %d \n", zab.Epoch)
	// 还没有开始投票， 投自己一票
	if zab.VoteMap[zab.ServerId] == nil {
		//调整epoch
		zab.IncrementEpoch()
		zab.VoteMap[zab.ServerId] = &Vote{
			VotedLeaderServerId: zab.ServerId,
			VotedLeaderZxId:     zab.ZxId(),
			State:               zab.State,
		}
	}
	//进行通知
	zab.notification()
	time.Sleep(time.Duration(rand.Int63()%300+510) * time.Millisecond)
	//检查状态
	zab.checkState()
}
func (zab *Zab) SendLog(log AppendLog, result *AppendLogResult) error {
	//防止服务端重复发送
	if len(log.Entries) > 0 && log.Entries[0].ZxId > zab.LastReceiveZxId {
		//这里简单直接追加日志
		zab.Logs = append(zab.Logs, log.Entries...)
		zab.LastReceiveZxId = log.Entries[len(log.Entries)-1].ZxId
		fmt.Println(log.Entries)
	}
	result.Ack = true
	return nil
}
func (zab *Zab) doSendLog(lastReceiveId uint64, serverId int, log []LogEntry) {
	client := getRpcClientWithReConn(zab, serverId)
	if client == nil {
		return
	}
	//防止重复发送，进行更改
	zab.ReceiveIdLock.Lock()
	zab.LastReceiveZxIdMap[serverId] = log[len(log)-1].ZxId
	zab.ReceiveIdLock.Unlock()
	r := new(AppendLogResult)
	err := client.SendLog(AppendLog{log}, r)
	//状态还原，进行重试
	if err != nil || !r.Ack {
		zab.ReceiveIdLock.Lock()
		zab.LastReceiveZxIdMap[serverId] = lastReceiveId
		zab.ReceiveIdLock.Unlock()
	}
	//网络有故障
	if err != nil {
		setRpcClient(zab, serverId, nil)
	}

}
func (zab *Zab) sendLog() {
	//更新已经提交
	commitZxId := zab.LastCommitZxId
	var zxids []uint64
	for _, v := range zab.LastReceiveZxIdMap {
		if v > commitZxId {
			zxids = append(zxids, v)
		}
	}
	f := func(a, b int) bool {
		return zxids[a] < zxids[b]
	}
	//说明有必要更新commitIndex
	if len(zxids) >= len(zab.OtherNodeAddress)/2 {
		sort.Slice(zxids, f)
		//最小的作为更新项目
		zab.LastCommitZxId = zxids[0]
	}
	//将日志发送到每个节点上
	logLen := len(zab.Logs)
	for k, _ := range zab.OtherNodeAddress {
		lastReceivedZxId := zab.LastReceiveZxIdMap[k]
		if lastReceivedZxId >= 0 {
			//选择符合要求的数据，发送给Follower节点，这里为了处理方便，全量的发送
			for i := 0; i < logLen; i++ {
				l := zab.Logs[i]
				if l.ZxId > lastReceivedZxId {
					//暂时先同步发
					func(lastReceiveZxId uint64, serverId int, log []LogEntry) {
						zab.doSendLog(lastReceivedZxId, serverId, log)
					}(lastReceivedZxId, k, zab.Logs[i:logLen])
					break
				}
			}
		}
	}
}

// CouldBroadCast 重新选举的leader需要在崩溃恢复后才能进行广播
func (zab *Zab) CouldBroadCast() bool {
	//已经被重置了，可以广播
	if couldBroadCast.Load() {
		return true
	}
	l := len(zab.Logs)
	if l == 0 {
		couldBroadCast.Store(true)
		return true
	}
	//已经提交了，可以进行
	if zab.Logs[l-1].ZxId < zab.LastCommitZxId {
		couldBroadCast.Store(true)
		return true
	}
	return false

}

// 是否接收到leader的心跳
var receivedHeatBeat atomic.Bool

func Follower(zab *Zab) {
	fmt.Printf("Follower EPOCH:%d \n", zab.Epoch)
	//重置计数器
	receivedHeatBeat.Store(false)
	time.Sleep(time.Duration(rand.Int63()%333+550) * time.Millisecond)
	//表示没有接收到请求， 转换为 LOOKING
	if !receivedHeatBeat.Load() {
		zab.StateLock.Lock()
		//重置投票箱子
		zab.resetVote()
		//进入选举状态
		zab.SetAction(Looking, LOOKING)
		zab.StateLock.Unlock()
	}
}
