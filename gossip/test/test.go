package main

import (
	"common"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	nodes := flag.Int("nodes", -1, "节点数量")
	port := flag.Int("port", 8080, "起始端口")
	iface := flag.String("iface", "WLAN", "网卡名称")
	flag.Parse()
	if *nodes == -1 {
		fmt.Println("缺少节点数量: -nodes")
		os.Exit(1)
	}

	portNumber := *port
	address := []string{}
	for i := 0; i < *nodes; i++ {
		address = append(address, ":"+strconv.Itoa(portNumber))
		portNumber++
	}
	fmt.Println(address)
	for i := 0; i < *nodes; i++ {
		go func(a int) {
			node := common.NewGossipNode(a, address[a], *iface)
			go node.Start()
		}(i)
	}
	time.Sleep(999999 * time.Second)

}
