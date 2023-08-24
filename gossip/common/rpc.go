package common

import "net/rpc"

type GossipRpc interface {
	Push(req map[string]*Pair, resp *struct{}) error
	Pull(req *struct{}, resp *struct{}) error

	Update(req Pair, resp *struct{}) error
	Get(key string, resp *Pair) error
}

type GossipRpcClient struct {
	*rpc.Client
	Target string
}

func (c GossipRpcClient) Push(req map[string]*Pair, resp *struct{}) error {
	return c.Client.Call(c.Target+".Push", req, resp)
}
func (c GossipRpcClient) Pull(req *struct{}, resp *struct{}) error {
	return c.Client.Call(c.Target+".Pull", req, resp)
}

func (c GossipRpcClient) Update(req Pair, resp *struct{}) error {
	return c.Client.Call(c.Target+".Update", req, resp)
}
func (c GossipRpcClient) Get(key string, resp *Pair) error {
	return c.Client.Call(c.Target+".Get", key, resp)
}
