package main

import (
	"common"
	"flag"
	"fmt"
	"os"
	"strconv"
)

func main() {
	nodes := flag.Int("nodes", -1, "节点数量")
	port := flag.Int("port", 8000, "起始端口")
	cur := flag.Int("cur", -1, "当前节点下标")
	flag.Parse()

	if *nodes == -1 {
		fmt.Println("缺少节点数量: -nodes")
		os.Exit(1)
	}
	if *cur == -1 {
		fmt.Println("节点下标小于0: -cur")
		os.Exit(1)
	}
	startRaft(*nodes, *port, *cur)
}

// 设置初始化的客户端的数量
func startRaft(nodes int, port int, cur int) {
	allAddress := map[int]string{}
	zab := common.NewZab(cur)
	i := 0
	for i < nodes {
		allAddress[i] = ":" + strconv.Itoa(port)
		port++
		i++
	}
	initAddress(zab, cur, allAddress)
	zab.StartListen()
	zab.InitRpcClient()
	zab.StartMainLoop()
}
func initAddress(r *common.Zab, index int, allAddress map[int]string) {
	otherAddress := make(map[int]string)
	for i, v := range allAddress {
		if i != index {
			otherAddress[i] = v
		}
	}
	fmt.Println("current address", allAddress[index])
	fmt.Println("other address:", otherAddress)
	r.Address = allAddress[index]
	r.OtherNodeAddress = otherAddress
}
