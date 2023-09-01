package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/pj"
	"github.com/vinceanalytics/vince/internal/render"
	"google.golang.org/protobuf/proto"
)

var client = &http.Client{}

func Apply(w http.ResponseWriter, r *http.Request) {
	log := slog.Default().With("endpoint", "Apply")
	var b v1.Cluster_Apply_Request
	err := pj.UnmarshalDefault(&b, r.Body)
	if err != nil {
		render.ERROR(w, http.StatusBadRequest)
		return
	}
	if b.Config == nil {
		render.ERROR(w, http.StatusBadRequest, "config is required")
		return
	}
	ctx := r.Context()

	for node, n := range b.Config.Nodes {
		// make sure we can reach all the nodes in this cluster.
		if f := ping(ctx, log, node, n); f != "" {
			render.ERROR(w, http.StatusBadRequest, f)
			return
		}
	}
	o := config.Get(ctx)
	var node *v1.Cluster_Config_Node
	for name, n := range b.Config.Nodes {
		if name == o.ServerId {
			node = n
			break
		}
	}
	if node == nil {
		render.ERROR(w, http.StatusBadRequest, "node is not part of the cluster")
		return
	}
	if !node.Bootstrap {
		render.ERROR(w, http.StatusBadRequest, "applying cluster config to non leader")
		return
	}
	db.Get(ctx).Txn(true, func(txn db.Txn) error {
		key := keys.Cluster()
		defer key.Release()
		return txn.Set(key.Bytes(),
			must.Must(proto.Marshal(b.Config))("failed serializing cluster config"),
		)
	})
	render.JSON(w, http.StatusOK, &v1.Cluster_Apply_Response{
		Ok: "applied",
	})
}

func Cluster(w http.ResponseWriter, r *http.Request) {
	var b v1.Cluster_Get_Request
	err := pj.UnmarshalDefault(&b, r.Body)
	if err != nil {
		render.ERROR(w, http.StatusBadRequest)
		return
	}
	var o v1.Cluster_Config
	db.Get(r.Context()).Txn(false, func(txn db.Txn) error {
		key := keys.Cluster()
		defer key.Release()
		return txn.Get(key.Bytes(), func(val []byte) error {
			return proto.Unmarshal(val, &o)
		})
	})
	render.JSON(w, http.StatusOK, &v1.Cluster_Get_Response{
		Config: &o,
	})
}

func ping(ctx context.Context, log *slog.Logger, name string, node *v1.Cluster_Config_Node) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, node.Address+"/cluster", nil)
	if err != nil {
		log.Error("failed creating ping  request for node",
			"node", name,
			"err", err.Error())
		return "invalid cluster configuration"
	}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("node:%q is not reachable", name)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusUnauthorized {
			return fmt.Sprintf("node:%q has invalid token", name)
		}
		log.Error("unexpected status code while pinging node",
			"node", name,
			"status", res.StatusCode,
		)
		return "invalid cluster configuration"
	}
	return ""
}
