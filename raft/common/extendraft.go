package common

import "fmt"

// 存放 raft 功能扩展相关的内容

type LogCommand struct {
	Operate string
	Key     string
	Value   string
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
	fmt.Println("Append 追加日志")
	r.Lock.Unlock()
	return nil
}
