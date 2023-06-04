package control

import (
	"context"
	"time"

	"github.com/vinceanalytics/vince/pkg/log"
	vince_informers "github.com/vinceanalytics/vince/v8s/gen/client/vince/informers/externalversions"
	vince_listers "github.com/vinceanalytics/vince/v8s/gen/client/vince/listers/vince/v1alpha1"
	"github.com/vinceanalytics/vince/v8s/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Options struct {
	MasterURL        string
	KubeConfig       string
	Port             int
	Image            string
	Namespace        string
	WatchNamespaces  []string
	IgnoreNamespaces []string
}

type Control struct {
	opts   Options
	queue  workqueue.RateLimitingInterface
	form   Inform
	list   List
	filter *k8s.ResourceFilter
	ready  func()
	top    *Topology
}

func New(clients k8s.Client, o Options, ready func()) *Control {
	x := Control{
		ready: ready,
		opts:  o,
		filter: k8s.NewResourceFilter(
			k8s.WatchNamespaces(o.WatchNamespaces...),
			k8s.IgnoreNamespaces(o.IgnoreNamespaces...),
			k8s.IgnoreNamespaces(metav1.NamespaceSystem),
		),
		form: Inform{
			vince: vince_informers.NewSharedInformerFactory(clients.Vince(), k8s.ResyncPeriod),
			k8s:   informers.NewSharedInformerFactory(clients.Kube(), k8s.ResyncPeriod),
		},
	}
	x.queue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	handler := cache.FilteringResourceEventHandler{
		FilterFunc: x.isWatchedResource,
		Handler:    &enqueueWorkHandler{queue: x.queue},
	}
	x.list.site = x.form.vince.Staples().V1alpha1().Sites().Lister()
	x.list.vince = x.form.vince.Staples().V1alpha1().Vinces().Lister()
	x.form.vince.Staples().V1alpha1().Sites().Informer().AddEventHandler(handler)
	x.form.vince.Staples().V1alpha1().Vinces().Informer().AddEventHandler(handler)
	x.top = NewTopology(
		clients,
		x.form.vince.Staples().V1alpha1().Vinces().Lister(),
		x.form.vince.Staples().V1alpha1().Sites().Lister(),
		x.form.k8s.Apps().V1().StatefulSets().Lister(),
		x.form.k8s.Core().V1().Services().Lister(),
		x.form.k8s.Core().V1().Secrets().Lister(),
	)
	return &x
}

type Inform struct {
	vince vince_informers.SharedInformerFactory
	k8s   informers.SharedInformerFactory
}

type List struct {
	site  vince_listers.SiteLister
	vince vince_listers.VinceLister
}

func (c *Control) Run(ctx context.Context) error {
	log.Get().Debug().Msg("Initializing vince controller")
	// we only watch Site resources
	{
		timeout, _ := context.WithTimeout(context.Background(), 10*time.Second)
		log.Get().Debug().Msg("Starting sites informer")
		c.form.vince.Start(ctx.Done())
		for _, ok := range c.form.vince.WaitForCacheSync(timeout.Done()) {
			if !ok {
				log.Get().Fatal().Msg("timed out waiting for controller caches to sync:")
			}
		}
	}
	c.ready()
	log.Get().Debug().Msg("Controller is ready")
	wait.Until(c.runWorker, time.Second, ctx.Done())
	return nil
}

// isWatchedResource returns true if the given resource is not ignored, false otherwise.
func (c *Control) isWatchedResource(obj interface{}) bool {
	return !c.filter.IsIgnored(obj)
}

func (c *Control) process() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)
	c.queue.Forget(key)
	return true
}

func (c *Control) runWorker() {
	for c.process() {
	}
}
