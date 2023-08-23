package common

import "net/rpc"

type GossipRpc interface {
	Push(req *struct{}, resp *struct{}) error
	Pull(req *struct{}, resp *struct{}) error
}

type GossipRpcClient struct {
	*rpc.Client
	// 被调用的方法路径
	Target string
}

func (c GossipRpcClient) Push(req *struct{}, resp *struct{}) error {
	return c.Client.Call(c.Target+".Push", req, resp)
}
func (c GossipRpcClient) Pull(req *struct{}, resp *struct{}) error {
	return c.Client.Call(c.Target+".Pull", req, resp)
}
