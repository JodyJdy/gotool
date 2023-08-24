package common

import (
	"net"
	"strconv"
	"strings"
)

type AddressWithPort struct {
	Ip   net.IP
	Port int
}

func GetRpcName(node int) string {
	return "GossipNode-" + strconv.Itoa(node)
}

func Parse(addressWithPort string) *AddressWithPort {
	result := new(AddressWithPort)
	splits := strings.Split(addressWithPort, ":")
	port, _ := strconv.Atoi(splits[1])
	result.Port = port
	ip := splits[0]
	ipslits := strings.Split(ip, ".")
	var bytes []byte
	for _, s := range ipslits {
		b, _ := strconv.Atoi(s)
		bytes = append(bytes, byte(b))
	}
	result.Ip = net.IPv4(bytes[0], bytes[1], bytes[2], bytes[3])
	return result
}
