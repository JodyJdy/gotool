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

type ZabError struct {
	msg string
}

func NewError(msg string) *ZabError {
	return &ZabError{msg: msg}
}
func (z *ZabError) Error() string {
	return z.msg
}

// AddLog 向Leader 添加一个日志
func (zab *Zab) AddLog(logCommand LogCommand, reply *struct{}) error {
	if zab.State != LEADER {
		return nil
	}
	if !zab.CouldBroadCast() {
		fmt.Println("崩溃恢复中，暂停广播")
	}
	zab.LogLock.Lock()
	defer zab.LogLock.Unlock()
	//添加计数器
	zab.IncrementCounter()
	log := LogEntry{
		zab.ZxId(), logCommand,
	}
	if zab.SendBroadCast(log) {
		//广播成功才追加到日志里面去
		zab.Logs = append(zab.Logs, log)
	} else {
		return NewError("消息广播失败")
	}
	return nil
}
