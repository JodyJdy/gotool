package main

import (
	"net/rpc"
	"test/base"
)

func main() {
	MasterAddress := "127.0.0.1:1234"
	MasterExposeRpcName := "Master"
	client, _ := rpc.Dial("tcp", MasterAddress)
	rpcClient := base.RpcClient{
		Client: client,
		Target: MasterExposeRpcName,
	}
	masterRpcClient := base.MasterRpcClient{
		RpcClient: rpcClient,
	}
	task := base.Task{
		TaskId:         base.GetTaskId(),
		ReducePhaseNum: 3,
		Files: []base.FileWithLocation{
			{
				FileName: "a.txt",
				NodeName: "master",
			},
			{
				FileName: "b.txt",
				NodeName: "master",
			},
			{
				FileName: "c.txt",
				NodeName: "master",
			},
		},
	}
	// 发布任务
	masterRpcClient.AddTask(task)
}
