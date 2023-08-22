package common

import (
	"encoding/gob"
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

// AppendLog 追加日志
func (r *Raft) AppendLog(logCommand LogCommand, reply *struct{}) error {
	if r.State != LEADER {
		return nil
	}
	r.Lock.Lock()
	r.Logs = append(r.Logs, LogEntry{
		r.getLastIndex() + 1, r.CurrentTerm, logCommand,
	})
	r.Lock.Unlock()
	return nil
}
