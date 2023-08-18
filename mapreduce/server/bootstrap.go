package main

import (
	"net"
	"net/rpc"
	"test/base"
)

func main() {

	// 初始化 master结构
	master := base.Master{
		Address:       ":8888",
		Workers:       []base.Worker{},
		FileOperate:   base.FileOperate{},
		ExposeRpcName: "Master",
		ClientMap:     make(map[base.Worker]*base.WorkerRpcClient),
	}
	//注册rpc服务
	rpc.RegisterName(master.ExposeRpcName, &master)
	listener, _ := net.Listen("tcp", master.Address)
	for {
		conn, _ := listener.Accept()
		go rpc.ServeConn(conn)
	}
}
