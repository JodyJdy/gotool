package main

import (
	"common"
)

func main() {

	raft := common.NewRaft()
	raft.Id = 4
	initAddress(raft, 3)
	raft.StartListen()
	//time.Sleep(10 * time.Second)
	raft.InitRpcClient()
	raft.StartMainLoop()
}
func initAddress(r *common.Raft, index int) {
	allAddress := []string{":1234", ":1235", ":1236", ":1237", ":1238"}
	var otherAddress []string
	for i, v := range allAddress {
		if i != index {
			otherAddress = append(otherAddress, v)
		}
	}
	r.Address = allAddress[index]
	r.OtherNodeAddress = otherAddress
}
