package main

import (
	"common"
	"time"
)

func main() {
	address := []string{
		":8080", ":8081", ":8082", ":8083", ":8084", ":8085", ":8086", ":8087", ":8088",
	}
	for i := 0; i < 9; i++ {
		go func(a int) {
			node := common.NewGossipNode(a, address[a])
			go node.Start()
		}(i)
	}
	time.Sleep(999999 * time.Second)

}
