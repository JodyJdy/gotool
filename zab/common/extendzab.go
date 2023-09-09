package common

import (
	"encoding/gob"
	"fmt"
)

// 存放 raft 功能扩展相关的内容

// LogCommand 日志命令定义
type LogCommand struct {
	Operate string
	Key     string
	Value   string
}

// 将该类型注册到gob中，这样远程调用就不会出错
func init() {
	gob.RegisterName("common.LogCommand", LogCommand{})
}

// AddLog 向Leader 添加一个日志
func (zab *Zab) AddLog(logCommand LogCommand, reply *struct{}) error {
	fmt.Println("添加日志")
	if zab.State != LEADER {
		return nil
	}
	if !zab.CouldBroadCast() {
		fmt.Println("崩溃恢复中，暂停广播")
	}

	zab.LogLock.Lock()
	//添加计数器
	zab.IncrementCounter()
	zab.Logs = append(zab.Logs, LogEntry{
		zab.ZxId(), logCommand,
	})
	zab.LogLock.Unlock()
	return nil
}
