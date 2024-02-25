package http

import (
	"context"
	"sort"
	"time"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

// NewNodesFromServers creates a slice of Nodes from a slice of Servers.
func NewNodesFromServers(servers *v1.Server_List) *v1.Node_List {
	nodes := &v1.Node_List{
		Items: make([]*v1.Node, len(servers.Items)),
	}
	for i, s := range servers.Items {
		nodes.Items[i] = &v1.Node{
			Id:    s.Id,
			Addr:  s.Addr,
			Voter: s.Suffrage == v1.Server_Voter,
		}
	}
	ls := nodes.Items
	sort.Slice(ls, func(i, j int) bool {
		return ls[i].Id < ls[j].Id
	})
	return nodes
}

// Test tests the node's reachability and leadership status. If an error
// occurs, the Error field will be populated.
func TestNode(ctx context.Context, n *v1.Node, ga GetAddresser, leaderAddr string, timeout time.Duration) {
	start := time.Now()
	n.Time = time.Since(start).Seconds()
	n.TimeS = time.Since(start).String()
	n.Reachable = false
	n.Leader = false
	doCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	apiAddr, err := ga.GetNodeAPIAddr(doCtx, n.Addr)
	if err != nil {
		n.Error = err.Error()
		return
	}
	n.ApiAddr = apiAddr
	n.Reachable = true
	n.Leader = n.Addr == leaderAddr
}

func Voters(n *v1.Node_List) (o *v1.Node_List) {
	o = &v1.Node_List{}
	for _, e := range n.Items {
		if e.Voter {
			o.Items = append(o.Items, e)
		}
	}
	ls := o.Items
	sort.Slice(ls, func(i, j int) bool {
		return ls[i].Id < ls[j].Id
	})
	return
}
