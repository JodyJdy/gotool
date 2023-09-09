package common

import "net/rpc"

type ZabRpc interface {
}

type ZabRpcClient struct {
	*rpc.Client
	// 被调用的方法路径
	Target string
}
