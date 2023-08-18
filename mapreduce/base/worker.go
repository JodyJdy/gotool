package base

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
	MasterRpcClient
}

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

func (w *Worker) MapTask(args MapTaskArgs, _ *struct{}) error {
	file := args.File
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
		w.MasterRpcClient.GetFile(file.FileName, bytes)
		pairSource = &BytesPairSource{
			fileName: file.FileName,
			contents: *bytes,
		}
	} else {
		//从特定节点获取数据
		worker := new(Worker)
		w.MasterRpcClient.GetWorker(file.NodeName, worker)
		if worker != nil {
			client := CreateWorkerMapClient(*worker)
			bytes := new([]byte)
			w.MasterRpcClient.GetFile(file.FileName, bytes)
			client.GetFile(file.FileName, bytes)
			pairSource = &BytesPairSource{
				fileName: file.FileName,
				contents: *bytes,
			}
		}
	}
	Map(pairSource, args.ReducePhaseNum, args.MapIndex, MapFunc)
	return nil
}
