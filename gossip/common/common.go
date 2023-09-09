package common

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"golang.org/x/net/ipv4"
	"math/rand"
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

// Pair 一个键值对
type Pair struct {
	// 键
	Key string
	// 值
	Value string
	// 数据版本号
	Version uint32
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
	// 记录集群节点的Client
	ClientMap map[int]*GossipRpcClient
	// 扇出数，发送数据时，随机选取FanOut个节点进行发送
	FanOut int
	// 节点状态信息
	MetaData map[string]*Pair
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
					//fmt.Printf("节点%d 接收到节点:%v \n", g.Id, node)
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
	node.MetaData = make(map[string]*Pair)
	node.ClientMap = make(map[int]*GossipRpcClient)
	return node
}

// StartRpcListen 开启Rpc监听
func (g *GossipNode) StartRpcListen() {
	rpc.RegisterName(GetRpcName(g.Id), g)
	listener, err := net.Listen("tcp", g.RpcAddress)
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			conn, _ := listener.Accept()
			go rpc.ServeConn(conn)
		}
	}()
	fmt.Printf("%d 开启Rpc监听\n", g.Id)
}

// StartAntiEntrop 开启反熵行为
func (g *GossipNode) StartAntiEntrop() {
	go func() {
		for {
			// 发送数据
			time.Sleep(1 * time.Second)
			// 随机选取节点
			var nodes []int
			for k, v := range g.NodeMessage {
				if v != nil {
					nodes = append(nodes, k)
				}
			}
			NodeNum := len(nodes)
			//集群中还没有节点注册
			if NodeNum == 0 {
				continue
			}

			var sendNodes []int
			//全量
			if NodeNum >= g.FanOut {
				sendNodes = nodes
			} else {
				//随机选取
				choose := make(map[int]bool)
				count := 0
				for {
					i := rand.Intn(len(nodes))
					if !choose[nodes[i]] {
						choose[nodes[i]] = true
						sendNodes = append(sendNodes, nodes[i])
						count++
					}
					if count == g.FanOut {
						break
					}
				}
			}
			// 发送数据
			for _, v := range sendNodes {
				go func(n int) {
					if g.ClientMap[n] == nil {
						g.lock.Lock()
						g.ClientMap[n] = g.createRpcClient(n)
						g.lock.Unlock()
					}
					if g.ClientMap[n] != nil {
						g.ClientMap[n].Push(g.MetaData, new(struct{}))
					}
				}(v)
			}

		}
	}()
}
func (g *GossipNode) createRpcClient(node int) *GossipRpcClient {
	message := g.NodeMessage[node]
	c, _ := rpc.Dial("tcp", message.RpcAddress)
	return &GossipRpcClient{Client: c, Target: GetRpcName(node)}
}

func (g *GossipNode) Start() {
	g.StartRpcListen()
	g.StartMembershipListen()
	// 要保证 第一次保活后，rpc监听已经打开
	g.StartKeepAlive()
	// 开启反熵策略
	g.StartAntiEntrop()
}

func (g *GossipNode) Push(req map[string]*Pair, resp *struct{}) error {
	// 收到其他节点的信息
	g.lock.Lock()
	for k, v := range req {
		if v != nil {
			if g.MetaData[k] == nil {
				g.MetaData[k] = v
				fmt.Printf("节点: %d  %s 接收到数据更新: %v \n", g.Id, g.RpcAddress, v)
			} else {
				if g.MetaData[k].Version < v.Version {
					g.MetaData[k].Version = v.Version
					g.MetaData[k].Value = v.Value
					fmt.Printf("节点: %d  %s 接收到数据更新: %v \n", g.Id, g.RpcAddress, v)
				}
			}
		}
	}
	g.lock.Unlock()
	return nil
}
func (g *GossipNode) Pull(req *struct{}, resp *struct{}) error {
	// 只做基于push的同步
	return nil
}

func (g *GossipNode) Update(req Pair, resp *struct{}) error {
	// 接收到的数据，是不考虑版本号的
	fmt.Printf("节点: %d 接收到数据 %v  \n", g.Id, req)
	if g.MetaData[req.Key] == nil {
		req.Version = 1
		g.lock.Lock()
		g.MetaData[req.Key] = &req
		g.lock.Unlock()
	} else {
		g.lock.Lock()
		g.MetaData[req.Key].Version++
		g.MetaData[req.Key].Value = req.Value
		g.lock.Unlock()
	}
	return nil
}

type KeyNotExist struct {
	err string
}

func (k *KeyNotExist) Error() string {
	return k.err
}
func (g *GossipNode) Get(key string, resp *Pair) error {
	if g.MetaData[key] == nil {
		return &KeyNotExist{err: key + " 不存在"}
	}
	*resp = *g.MetaData[key]
	return nil
}
