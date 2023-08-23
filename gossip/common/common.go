package common

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"golang.org/x/net/ipv4"
	"net"
	"net/rpc"
	"sync"
	"time"
)

// Node 节点信息
type Node struct {
	//节点id
	Id int
	// 节点rpc地址
	RpcAddress string
}

type GossipNode struct {
	lock sync.Mutex
	// gossip集群 多播地址
	MultiCastAddress string
	// 当前节点id
	Id int
	// 当前节点 暴露rpc服务地址
	RpcAddress string
	// 节点下线时间，一定时间没有保活就认为节点失效
	NodeDownTime int
	//记录集群中的节点，以及节点上次保活的时间
	NodeStatus map[int]int
	// 记录集群节点的信息
	NodeMessage map[int]*Node
	// 扇出数，发送数据时，随机选取FanOut个节点进行发送
	FanOut int
}

func Server(addressWithPort *AddressWithPort) *net.UDPConn {
	ipv4Addr := &net.UDPAddr{IP: addressWithPort.Ip, Port: addressWithPort.Port}
	conn, err := net.ListenUDP("udp4", ipv4Addr)
	if err != nil {
		fmt.Printf("ListenUDP error %v\n", err)
		return nil
	}
	pc := ipv4.NewPacketConn(conn)
	// 根据网卡名称获取网卡
	iface, err := net.InterfaceByName("以太网")
	if err != nil {
		fmt.Printf("找不到网卡 %v\n", err)
		return nil
	}
	if err := pc.JoinGroup(iface, &net.UDPAddr{IP: addressWithPort.Ip}); err != nil {
		return nil
	}
	//服务启动
	return conn
}

// StartMembershipListen 打开集群成员消息的监听
func (g *GossipNode) StartMembershipListen() {
	conn := Server(Parse(g.MultiCastAddress))
	go func(con *net.UDPConn) {
		buf := make([]byte, 1024)
		for {
			n, _, _ := conn.ReadFrom(buf)
			buffer := bytes.NewBuffer(buf[:n])
			x := gob.NewDecoder(buffer)
			node := new(Node)
			if x.Decode(node) == nil {
				//忽略自身发送的消息
				if node.Id != g.Id {
					g.lock.Lock()
					g.NodeStatus[node.Id] = time.Now().Second()
					g.NodeMessage[node.Id] = node
					g.lock.Unlock()
					fmt.Printf("节点%d 接收到节点:%v \n", g.Id, node)
				}
			}
		}

	}(conn)
	fmt.Printf("%d 集群成员变更监听 启动\n", g.Id)
}
func Client(addressWithPort *AddressWithPort) *net.UDPConn {
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: addressWithPort.Ip, Port: addressWithPort.Port}
	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		fmt.Println(err)
	}
	return conn
}

// StartKeepAlive 开启自身的保活
func (g *GossipNode) StartKeepAlive() {
	conn := Client(Parse(g.MultiCastAddress))
	self := Node{
		Id:         g.Id,
		RpcAddress: g.RpcAddress,
	}
	go func(con *net.UDPConn, self Node) {
		buffer := bytes.Buffer{}
		x := gob.NewEncoder(&buffer)
		if x.Encode(self) != nil {
			panic("")
		}
		// 周期性的保活，并清除过期的节点
		for {
			_, r := conn.Write(buffer.Bytes())
			if r != nil {
				fmt.Println(r)
			}
			time.Sleep(5 * time.Second)
			g.lock.Lock()
			var deleteKey []int
			for k, v := range g.NodeStatus {
				// 节点下线
				if time.Now().Second()-v > g.NodeDownTime {
					deleteKey = append(deleteKey, k)
				}
			}
			for _, k := range deleteKey {
				fmt.Printf("节点 %d 中 下线: %d \n", g.Id, k)
				delete(g.NodeStatus, k)
				delete(g.NodeMessage, k)
			}
			g.lock.Unlock()
		}

	}(conn, self)
	fmt.Printf("%d 开启保活\n", g.Id)
}

func NewGossipNode(id int, rpcAddress string) *GossipNode {
	node := new(GossipNode)
	node.Id = id
	node.MultiCastAddress = "228.0.0.4:8899"
	node.NodeStatus = make(map[int]int)
	node.NodeMessage = make(map[int]*Node)
	node.NodeDownTime = 10
	node.RpcAddress = rpcAddress
	node.FanOut = 3
	return node
}

// StartRpcListen 开启Rpc监听
func (g *GossipNode) StartRpcListen() {
	rpc.RegisterName("GossipNode", g)
	listener, _ := net.Listen("tcp", g.RpcAddress)
	go func() {
		for {
			conn, _ := listener.Accept()
			go rpc.ServeConn(conn)
		}
	}()
	fmt.Printf("%d 开启Rpc监听\n", g.Id)
}

func (g *GossipNode) Start() {
	g.StartRpcListen()
	g.StartMembershipListen()
	// 要保证 第一次保活后，rpc监听已经打开
	g.StartKeepAlive()
}

func (g *GossipNode) Push(req *struct{}, resp *struct{}) error {
	return nil
}
func (g *GossipNode) Pull(req *struct{}, resp *struct{}) error {
	return nil
}
