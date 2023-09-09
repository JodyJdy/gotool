package common

import (
	"fmt"
	"net"
	"net/rpc"
	"reflect"
	"sync"
)

func getRpcClient(zab *Zab, serverId int) *ZabRpcClient {
	client, ok := zab.Peers.Load(serverId)
	if ok {
		return reflect.ValueOf(client).Interface().(*ZabRpcClient)
	}
	return nil
}
func setRpcClient(zab *Zab, serverId int, client *ZabRpcClient) {
	zab.Peers.Store(serverId, client)
}

// StartListen 开启监听
func (zab *Zab) StartListen() {
	// 注册服务
	rpc.RegisterName("Zab", zab)
	listener, _ := net.Listen("tcp", zab.Address)
	go func() {
		for {
			conn, _ := listener.Accept()
			go rpc.ServeConn(conn)
		}
	}()
}

func (zab *Zab) InitRpcClient() {
	zab.Peers = sync.Map{}
	for i, v := range zab.OtherNodeAddress {
		conn, err := net.Dial("tcp", v)
		fmt.Println(err)
		if err == nil {
			cli := rpc.NewClient(conn)
			setRpcClient(zab, i, &ZabRpcClient{
				Client: cli,
				Target: "Zab",
			})
		}
	}
}

func (rf *Zab) CheckAndResetConn() {
	for i, v := range rf.OtherNodeAddress {
		if getRpcClient(rf, i) == nil {
			conn, err := net.Dial("tcp", v)
			if err == nil {
				cli := rpc.NewClient(conn)
				setRpcClient(rf, i, &ZabRpcClient{
					Client: cli,
					Target: "Zab",
				})
			}
		}
	}
}

// ResetConn 重置某个连接
func (zab *Zab) ResetConn(peer int) {
	conn, err := net.Dial("tcp", zab.OtherNodeAddress[peer])
	if err == nil {
		cli := rpc.NewClient(conn)
		setRpcClient(zab, peer, &ZabRpcClient{
			Client: cli,
			Target: "Zab",
		})
	}
}

// StartMainLoop 执行 Zab主循环
func (zab *Zab) StartMainLoop() {
	for {
		//执行主循环
		zab.Action(zab)
	}
}

func Leader(zab *Zab) {
}
func Looking(zab *Zab) {

}

func Follower(zab *Zab) {

}
