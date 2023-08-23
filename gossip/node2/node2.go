package main

import (
	"common"
	"time"
)

func main() {
	node := common.NewGossipNode(2, ":1235")
	node.StartMembershipListen()
	node.StartKeepAlive()
	time.Sleep(999999 * time.Second)
}
