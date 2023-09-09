package main

import (
	"common"
	"net/rpc"
)

func main() {
	c, _ := rpc.Dial("tcp", ":8080")
	y := common.GossipRpcClient{Client: c, Target: common.GetRpcName(0)}
	y.Update(common.Pair{Key: "hello", Value: "workd"}, new(struct{}))
}
