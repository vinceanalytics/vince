package topology

import (
	vince_listers "github.com/gernest/vince/pkg/gen/client/vince/listers/vince/v1alpha1"
	"github.com/gernest/vince/pkg/k8s"
	listers "k8s.io/client-go/listers/core/v1"
)

type Topology struct {
	serviceLister listers.ServiceLister
	vinceLister   vince_listers.VinceLister
	podLister     listers.PodLister
	filter        *k8s.ResourceFilter
}
