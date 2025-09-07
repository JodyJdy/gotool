package main

import (
	"common"
	"flag"
	"fmt"
	"net/rpc"
	"os"
	"strconv"
	"time"
)

func main() {
	cur := flag.Int("cur", -1, "发送节点下标")
	port := flag.Int("port", 8080, "起始端口")
	flag.Parse()
	if *cur == -1 {
		fmt.Println("缺少发送节点的下标: -cur")
		os.Exit(1)
	}
	nodePort := *port + (*cur)
	fmt.Println("发送到的节点:", nodePort)
	addr := ":" + strconv.Itoa(nodePort)
	c, _ := rpc.Dial("tcp", addr)
	y := common.GossipRpcClient{Client: c, Target: common.GetRpcName(*cur)}
	i := 0
	for {
		y.Update(common.Pair{Key: "hello", Value: "world" + strconv.Itoa(i)}, new(struct{}))
		time.Sleep(5 * time.Second)
		i++
	}
}
