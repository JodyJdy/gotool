package main

import (
	"common"
	"time"
)

func main() {
	address := []string{
		":1234", ":1235", ":1236", ":1237", ":1238",
	}
	for i := 1; i < 5; i++ {
		node := common.NewGossipNode(i, address[i-1])
		node.Start()
	}
	time.Sleep(999999 * time.Second)

}
