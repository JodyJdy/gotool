package main

import (
	"common"
	"fmt"
	"net"
	"net/rpc"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", ":1234")
	if err == nil {
		cli := rpc.NewClient(conn)

		x := &common.RaftRpcClient{
			Client: cli,
			Target: "Raft",
		}
		for {
			err := x.AppendLog(common.LogCommand{"add", "hello", "world"}, new(struct{}))
			fmt.Println(err)
			time.Sleep(5 * time.Second)
		}
	}
}
