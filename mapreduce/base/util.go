package base

import (
	"fmt"
	"math/rand"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// GetTaskId 生成任务id
func GetTaskId() string {
	return RandStringBytes(10)
}

// GetIntermediateFileName 获取map后生成的中间文件名称
func GetIntermediateFileName(taskId string, reduceIndex int, mapIndex int) string {
	return fmt.Sprintf("%s-intermediate-%d-%d", taskId, reduceIndex, mapIndex)
}

// GetOutputFileName 获取最终的输出文件名称
func GetOutputFileName(taskId string, reduceIndex int) string {
	return fmt.Sprintf("%soutput-%d", taskId, reduceIndex)
}
