package base

import (
	"log"
)

// Map nReduce 映射为n个中间文件
func Map(source PairSource, nReduce int, mapTaskIndex int, mapF func(p Pair) []Pair) {

	// 获取输入 mapF 的 key-value对
	pair := source.ReadPair()
	// 执行函数,生产 key-value对 列表， 用作 reduce的输入
	pairsResult := mapF(pair)
	//持久化
	log.Fatal(pairsResult)

}
