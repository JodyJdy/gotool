package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {
	if len(os.Args) > 1 {
		for _, s := range os.Args[1:] {
			splits := strings.Split(s, "-")
			AddPortMapping(splits[1], splits[0])
		}
	}
	time.Sleep(99999 * time.Second)

}

// 全局锁
var lock sync.Mutex

// 全局映射信息
var portMapping *PortMapping

// PortMapping 记录端口映射的所有信息
type PortMapping struct {
	// 记录端口以及端口映射地址
	TargetWithPort map[string][]string
	// 记录端口是否映射
	MappingStatus map[string]bool
	// 映射多个地址时，记录负载均衡的信息
	LoadBalancer map[string]int
}

func (p *PortMapping) addPortMapping(toAddr string, fromAddr string) {
	fmt.Printf("添加端口映射: %s-%s\n", fromAddr, toAddr)
	status := p.MappingStatus[fromAddr]
	lock.Lock()
	p.TargetWithPort[fromAddr] = append(p.TargetWithPort[fromAddr], toAddr)
	p.MappingStatus[fromAddr] = true
	lock.Unlock()

	// 第一次添加该端口的映射
	if !status {
		go listen(fromAddr)
	}
}
func (p *PortMapping) removePortMapping(fromAddr string, toAddr string) {
	lock.Lock()
	s := p.TargetWithPort[fromAddr]
	var result []string
	for _, v := range s {
		// 去掉addr
		if v != toAddr {
			result = append(result, v)
		}
	}
	p.TargetWithPort[fromAddr] = result
	if len(result) == 0 {
		p.MappingStatus[fromAddr] = false
	}
	lock.Unlock()
}
func (p *PortMapping) removePort(fromAddr string) {
	lock.Lock()
	p.TargetWithPort[fromAddr] = []string{}
	p.MappingStatus[fromAddr] = false
	lock.Unlock()
}

// 负载均衡
func (p *PortMapping) loadBalancer(fromAddr string) string {
	// 端口映射关闭，不再接收连接
	if !p.MappingStatus[fromAddr] {
		return ""
	}
	lock.Lock()
	defer lock.Unlock()
	p.LoadBalancer[fromAddr] = (p.LoadBalancer[fromAddr] + 1) % len(p.TargetWithPort[fromAddr])
	return p.TargetWithPort[fromAddr][p.LoadBalancer[fromAddr]]
}
func listen(fromAddr string) {
	fromlistener, err := net.Listen("tcp", fromAddr)
	if fromlistener == nil {
		return
	}
	defer fromlistener.Close()
	if err != nil {
		fmt.Printf("监听端口 %s, 失败: %s\n", fromAddr, err.Error())
		return
	}
	// 进入循环
	for portMapping.MappingStatus[fromAddr] {
		// 接收连接
		fromcon, err := fromlistener.Accept()
		//这边最好也做个协程，防止阻塞
		toAddr := portMapping.loadBalancer(fromAddr)
		log.Printf("接收连接  fro:%s to %s\n", fromAddr, toAddr)
		fmt.Printf("接收连接  fro:%s to %s\n", fromAddr, toAddr)
		toCon, err := net.Dial("tcp", toAddr)
		if err != nil {
			fmt.Printf("不能连接到 %s\n", toAddr)
			fromcon.Close()
			continue
		}
		go io.Copy(fromcon, toCon)
		go io.Copy(toCon, fromcon)
	}
}

func AddPortMapping(toAddr string, fromAddr string) {
	// 如果没有执行初始化，执行初始化
	if portMapping == nil {
		portMapping = initPortMapping()
	}
	portMapping.addPortMapping(toAddr, fromAddr)
}
func initPortMapping() *PortMapping {
	return &PortMapping{
		TargetWithPort: make(map[string][]string),
		MappingStatus:  make(map[string]bool),
		LoadBalancer:   make(map[string]int),
	}
}
