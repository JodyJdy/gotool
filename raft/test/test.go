package main

import (
	"common"
	"fmt"
	"runtime"
	"strconv"
	"time"
)

func main() {
	runtime.GOMAXPROCS(20)
	testRaft(3)
	time.Sleep(10000 * time.Second)
}

// 设置初始化的客户端的数量
func testRaft(clientNum int) {
	i := 1
	portNumber := 8000
	allAddress := []string{}
	raft_array := []*common.Raft{}
	for i <= clientNum {
		raft := common.NewRaft()
		raft.Id = i
		raft_array = append(raft_array, raft)
		allAddress = append(allAddress, ":"+strconv.Itoa(portNumber))
		portNumber++
		i++
	}
	for i, v := range raft_array {
		fmt.Println("node : ", i)
		initAddress(v, i, allAddress)
	}

	for _, v := range raft_array {
		v.StartListen()
	}
	for _, v := range raft_array {
		v.InitRpcClient()
	}
	for _, v := range raft_array {
		go v.StartMainLoop()
	}
}
func initAddress(r *common.Raft, index int, allAddress []string) {
	var otherAddress []string
	for i, v := range allAddress {
		if i != index {
			otherAddress = append(otherAddress, v)
		}
	}
	fmt.Println("other address:", otherAddress)
	r.Address = allAddress[index]
	r.OtherNodeAddress = otherAddress
}
