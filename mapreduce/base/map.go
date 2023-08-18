package base

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// Map nReduce 映射为n个中间文件
func Map(source PairSource, arg MapTaskArgs, mapF func(p Pair) []Pair) {

	// 获取输入 mapF 的 key-value对
	pair := source.ReadPair()
	// 执行函数,生产 key-value对 列表， 用作 reduce的输入
	pairsResult := mapF(pair)
	fmt.Println(pairsResult)

	// 存储中间文件
	var tmpFiles = make([]*os.File, arg.ReducePhaseNum)
	var encoders = make([]*json.Encoder, arg.ReducePhaseNum)

	for i := 0; i < arg.ReducePhaseNum; i++ {
		tmpFileName := GetIntermediateFileName(arg.TaskId, arg.MapIndex, i)
		tmpFiles[i], _ = os.Create(tmpFileName)
		fmt.Println("创建文件" + tmpFileName)
		defer tmpFiles[i].Close()
		encoders[i] = json.NewEncoder(tmpFiles[i])
	}

	for _, kv := range pairsResult {
		hashKey := int(kv.hashCode()) % arg.ReducePhaseNum
		err := encoders[hashKey].Encode(&kv)
		if err != nil {
			log.Fatal("do map encoders ", err)
		}
	}
}
