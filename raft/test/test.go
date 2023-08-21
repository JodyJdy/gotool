package main

import "common"

func main() {

	r := new(common.Raft)
	r.Action = common.LeaderFirst
	r.Action(r)
	r.Action(r)
}
