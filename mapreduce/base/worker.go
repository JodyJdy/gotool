package base

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

type Worker struct {
	// Worker名称，默认取WorkerAddress
	Name string
	// 集群节点地址
	MasterAddress string
	// 集群节点暴露的Rpc服务
	MasterExposeRpcName string
	// Worker 地址
	WorkerAddress string
	// Worker暴露的Rpc服务名称
	WorkerExposeRpcName string
}

var GlobalMasterRpcClient MasterRpcClient
var OtherWorkerRpcClient = make(map[string]*WorkerRpcClient)

// ShutDown 关闭Worker
func (w *Worker) ShutDown(_, _ *struct{}) error {
	return nil
}

// MapTaskArgs Map 阶段的参数
type MapTaskArgs struct {
	TaskId         string
	File           FileWithLocation
	MapIndex       int
	ReducePhaseNum int
}

// ReduceTaskArgs Reduce阶段的参数
type ReduceTaskArgs struct {
	TaskId           string
	ReducePhaseIndex int
	// Worker 完成的任务
	WorkerTaskMap map[string][]int
}

func (w *Worker) getPairSourceFromFile(file FileWithLocation) PairSource {
	// 处理文件
	var pairSource PairSource
	// 如果文件在当前节点，直接读取
	if file.NodeName == w.Name || file.IsLocal {
		pairSource = &FilePairSource{
			fileName: file.FileName,
		}
	} else if file.NodeName == MASTER_NODE_NAME {
		// 从主节点获取数据
		bytes := new([]byte)
		GlobalMasterRpcClient.GetFile(file.FileName, bytes)
		pairSource = &BytesPairSource{
			fileName: file.FileName,
			contents: *bytes,
		}
	} else {
		//从特定节点获取数据
		worker := new(Worker)
		GlobalMasterRpcClient.GetWorker(file.NodeName, worker)
		if worker != nil {
			client := CreateWorkerMapClient(*worker)
			bytes := new([]byte)
			GlobalMasterRpcClient.GetFile(file.FileName, bytes)
			client.GetFile(file.FileName, bytes)
			pairSource = &BytesPairSource{
				fileName: file.FileName,
				contents: *bytes,
			}
		}
	}
	return pairSource
}

func (w *Worker) MapTask(args MapTaskArgs, _ *struct{}) error {
	fmt.Println("处理文件")
	Map(w.getPairSourceFromFile(args.File), args, MapFunc)
	return nil
}

func (w *Worker) ReduceTask(args ReduceTaskArgs, _ *struct{}) error {
	var decoders []*json.Decoder
	for k, v := range args.WorkerTaskMap {
		//本地文件
		if k == w.Name {
			//读取文件
			for _, mapTaskIndex := range v {
				fileName := GetIntermediateFileName(args.TaskId, mapTaskIndex, args.ReducePhaseIndex)
				fmt.Println("读取文件" + fileName)
				file, _ := os.Open(fileName)
				decoders = append(decoders, json.NewDecoder(file))
			}

		} else {
			//创建向其他Worker的连接
			if OtherWorkerRpcClient[k] == nil {
				worker := new(Worker)
				GlobalMasterRpcClient.GetWorker(k, worker)
				OtherWorkerRpcClient[k] = CreateWorkerMapClient(*worker)
			}
			for _, mapTaskIndex := range v {
				fileName := GetIntermediateFileName(args.TaskId, mapTaskIndex, args.ReducePhaseIndex)
				con := new([]byte)
				OtherWorkerRpcClient[k].GetFile(fileName, con)
				decoders = append(decoders, json.NewDecoder(bytes.NewBuffer(*con)))
			}
		}

	}
	Reduce(decoders, args, ReduceFunc)
	return nil

}
