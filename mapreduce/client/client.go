package main

import (
	"net"
	"net/rpc"
	"test/base"
)

func main() {
	worker := base.Worker{
		Name:                "worker1",
		MasterAddress:       "127.0.0.1:1234",
		WorkerAddress:       "127.0.0.1:2234",
		MasterExposeRpcName: "Master",
		WorkerExposeRpcName: "Client",
	}

	//保证，监听启动后，才注册自身
	startListen := make(chan bool)
	//获取 master rpc client
	masterRpcClient := createMasterRpcClient(worker)
	worker.MasterRpcClient = masterRpcClient
	// 提前暴露worker的rpc服务
	go func() {
		rpc.RegisterName(worker.WorkerExposeRpcName, &worker)
		listener, _ := net.Listen("tcp", worker.WorkerAddress)
		startListen <- true
		for {
			conn, _ := listener.Accept()
			go rpc.ServeConn(conn)
		}
	}()
	//注册自身到Worker上面，要保证注册的时候，rpc服务已经暴露
	<-startListen
	masterRpcClient.RegisterWorker(worker)

	<-startListen

}
func createMasterRpcClient(worker base.Worker) base.MasterRpcClient {
	client, _ := rpc.Dial("tcp", worker.MasterAddress)
	rpcClient := base.RpcClient{
		Client: client,
		Target: worker.MasterExposeRpcName,
	}
	return base.MasterRpcClient{
		RpcClient: rpcClient,
	}
}
