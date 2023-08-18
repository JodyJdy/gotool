package base

const MASTER_NODE_NAME = "master"

type FileWithLocation struct {
	// 包含文件名称以及文件路径
	FileName string
	//所处的节点
	NodeName string
	// 本地文件
	IsLocal bool
}
type Task struct {
	// 任务Id
	TaskId string
	// ReducePhase阶段的任务数
	ReducePhaseNum int
	// 处理的所有文件
	Files []FileWithLocation
}
