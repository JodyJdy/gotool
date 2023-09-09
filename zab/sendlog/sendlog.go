package main

import (
	"common"
	"fmt"
	"net"
	"net/rpc"
	"time"
)

/*
*
测试日志的添加
*/
func main() {
	conn, err := net.Dial("tcp", ":1235")
	if err == nil {
		cli := rpc.NewClient(conn)

		x := &common.ZabRpcClient{
			Client: cli,
			Target: "Zab",
		}

		for {
			err := x.AddLog(common.LogCommand{"add", "hello", "world"}, new(struct{}))
			fmt.Println(err)
			time.Sleep(5 * time.Second)
		}
	}
}
