package main

import (
	"common"
	"flag"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"time"
)

func main() {
	leader := flag.String("leader", "", "leader地址")
	flag.Parse()
	if *leader == "" {
		fmt.Println("缺少 leader 地址: -leader")
		os.Exit(1)
	}
	conn, err := net.Dial("tcp", *leader)
	if err == nil {
		cli := rpc.NewClient(conn)
		x := &common.RaftRpcClient{
			Client: cli,
			Target: "Raft",
		}
		for {
			err := x.AppendLog(common.LogCommand{"add", "hello", "world"}, new(struct{}))
			if err != nil {
				fmt.Println(err)
				break
			}
			time.Sleep(5 * time.Second)
		}
	} else {
		fmt.Println("连接leader失败:", err)
	}
}
