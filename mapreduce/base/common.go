package base

import (
	"hash/fnv"
	"log"
	"net/rpc"
	"os"
)

type Value any

type Pair struct {
	Key string
	Val Value
}

// 计算HashCode
func (p Pair) hashCode() uint32 {
	h := fnv.New32a()
	h.Write([]byte(p.Key))
	return h.Sum32()
}

// PairSource 获取 Pair
type PairSource interface {
	ReadPair() Pair
}

type BytesPairSource struct {
	fileName string
	contents []byte
}

func (f *BytesPairSource) ReadPair() Pair {
	return Pair{
		Key: f.fileName,
		Val: f.contents,
	}
}

// FilePairSource 从文件中读取Pair
type FilePairSource struct {
	fileName string
}

func (f *FilePairSource) ReadPair() Pair {
	contents, _ := os.ReadFile(f.fileName)
	return Pair{
		Key: f.fileName,
		Val: contents,
	}
}

type FileRpc interface {
	// GetFile 读取远程主机上的文件， fileName 远程主机上的文件路径
	GetFile(fileName string, result *[]byte) error
	// SendFile 向远程主机发送文件
	SendFile(entity FileEntity, reply *string) error
}

type FileOperate struct {
}
type FileEntity struct {
	FileName string
	Contents []byte
}

func (*FileOperate) GetFile(fileName string, result *[]byte) error {
	if contents, err := os.ReadFile(fileName); err == nil {
		// 要修改指针对应的内存
		*result = append(*result, contents...)
		return nil
	} else {
		return err
	}
}
func (*FileOperate) SendFile(fileEntity FileEntity, reply *string) error {
	if file, err := os.Create(fileEntity.FileName); err != nil {
		log.Fatal(err)
		return err
	} else {
		_, _ = file.Write(fileEntity.Contents)
		file.Close()
		return nil
	}
}

type RpcServer interface {
	FileRpc
}

type RpcClient struct {
	*rpc.Client
	// 被调用的方法路径
	Target string
}

func (r RpcClient) SendFile(entity FileEntity, reply *string) error {
	return r.Client.Call(r.Target+".SendFile", entity, new(string))
}

func (r RpcClient) GetFile(fileName string, result *[]byte) error {
	return r.Client.Call(r.Target+".GetFile", fileName, result)
}

// MasterRpcClient worker 调用 Master使用
type MasterRpcClient struct {
	RpcClient
}

func (m *MasterRpcClient) RegisterWorker(worker Worker) error {
	client := m.RpcClient
	return client.Call(client.Target+".RegisterWorker", worker, new(struct{}))
}
func (m *MasterRpcClient) GetAllWorkers(workers *[]Worker) error {
	client := m.RpcClient
	return client.Call(client.Target+".GetAllWorkers", new(struct{}), workers)
}
func (m *MasterRpcClient) GetWorker(nodeName string, worker *Worker) error {
	client := m.RpcClient
	return client.Call(client.Target+".GetWorker", nodeName, worker)
}
func (m *MasterRpcClient) AddTask(task Task) error {
	client := m.RpcClient
	return client.Call(client.Target+".AddTask", task, new(struct{}))
}

// ClientRpcClient master调用 client
type ClientRpcClient struct {
	RpcClient
}

type WorkerRpcClient struct {
	RpcClient
}

func (w *WorkerRpcClient) ShutDown(_, _ *struct{}) error {
	client := w.RpcClient
	return client.Call(client.Target+".ShutDown", new(struct{}), new(struct{}))
}
func (w *WorkerRpcClient) MapTask(args MapTaskArgs) error {
	client := w.RpcClient
	return client.Call(client.Target+".MapTask", args, new(struct{}))
}
