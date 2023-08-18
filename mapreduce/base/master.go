package base

import (
	"fmt"
	"log"
	"net/rpc"
	"sync"
)

type Master struct {
	// 服务端地址
	Address string
	// 注册的多个Worker
	Workers []Worker
	// 暴露的Rpc服务名称
	ExposeRpcName string
	// 支持文件操作
	FileOperate
	// Worker -> WorkerRpcClient映射
	ClientMap map[Worker]*WorkerRpcClient
}

var lock sync.Mutex

// RegisterWorker 注册工作者
func (m *Master) RegisterWorker(worker Worker, _ *struct{}) error {
	// 会同时有多个Worker注册，考虑线程安全
	lock.Lock()
	defer lock.Unlock()
	log.Printf("接收到Worker注册: %s", worker)
	m.Workers = append(m.Workers, worker)
	m.ClientMap[worker] = CreateWorkerMapClient(worker)
	fmt.Println(m.ClientMap)
	return nil
}
func (m *Master) Hello(_ *struct{}, _ *struct{}) error {
	fmt.Println("fuck hello")
	return nil
}

// GetAllWorkers 获取所有的工作者
func (m *Master) GetAllWorkers(_ *struct{}, workers *[]Worker) error {
	*workers = append(*workers, m.Workers...)
	return nil
}

// GetWorker 获取指定的节点
func (m *Master) GetWorker(nodeName string, worker *Worker) error {
	for _, v := range m.Workers {
		if v.Name == nodeName {
			*worker = v
		}
	}
	return nil
}

// AddTask 启动一个任务
func (m *Master) AddTask(task Task, _ *struct{}) error {
	go m.ScheduleTask(task)
	return nil
}

// ScheduleTask 进行具体的调度
func (m *Master) ScheduleTask(task Task) {
	// 有几个文件 就执行几个 mapTask
	done := make(chan bool)
	// 第一阶段
	workerTaskMap := make(map[string][]int)
	// 轮训的将任务分给Worker
	fileNum := len(task.Files)
	for i := 0; i < fileNum; {
		for _, w := range m.Workers {
			if i >= fileNum {
				break
			}
			_, ok := workerTaskMap[w.Name]
			if !ok {
				workerTaskMap[w.Name] = []int{}
			}
			workerTaskMap[w.Name] = append(workerTaskMap[w.Name], i)
			args := MapTaskArgs{
				TaskId:         task.TaskId,
				File:           task.Files[i],
				MapIndex:       i,
				ReducePhaseNum: task.ReducePhaseNum,
			}
			go func(worker Worker) {
				m.ClientMap[worker].MapTask(args)
				done <- true
			}(w)
			i++
		}
	}
	//等待map执行完成
	for i := 0; i < fileNum; i++ {
		<-done
	}
	// 第二阶段
	//收集所有使用到的工作节点
	//执行 reduce 任务
	for i := 0; i < task.ReducePhaseNum; {
		for _, w := range m.Workers {
			if i >= task.ReducePhaseNum {
				break
			}
			args := ReduceTaskArgs{
				TaskId:           task.TaskId,
				ReducePhaseIndex: i,
				WorkerTaskMap:    workerTaskMap,
			}
			go func(worker Worker) {
				fmt.Println(args)
				m.ClientMap[worker].ReduceTask(args)
				done <- true
			}(w)
			i++
		}
	}
	// 等待reduce完成
	for i := 0; i < task.ReducePhaseNum; i++ {
		<-done
	}
	fmt.Printf("%s任务执行完毕", task.TaskId)
}

// CreateWorkerMapClient 由于 Worker在 注册前已经保证了服务的监听已经建立，这里可以直接创建client
func CreateWorkerMapClient(worker Worker) *WorkerRpcClient {
	client, _ := rpc.Dial("tcp", worker.WorkerAddress)
	c := WorkerRpcClient{
		RpcClient: RpcClient{
			Client: client,
			Target: worker.WorkerExposeRpcName,
		},
	}
	return &c
}
