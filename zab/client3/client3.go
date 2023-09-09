package main

import (
	"common"
)

func main() {
	zab := common.NewZab(3)
	initAddress(zab, zab.ServerId)
	zab.StartListen()
	zab.InitRpcClient()
	zab.StartMainLoop()
}

func initAddress(z *common.Zab, serverId int) {
	allAddress := map[int]string{1: ":1234", 2: ":1235", 3: ":1236", 4: ":1237", 5: ":1238"}
	otherAddress := make(map[int]string)
	for i, v := range allAddress {
		if i != serverId {
			otherAddress[i] = v
		}
	}
	z.Address = allAddress[serverId]
	z.OtherNodeAddress = otherAddress
}
