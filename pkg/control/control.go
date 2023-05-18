package control

import (
	"context"
	"time"

	vince_informers "github.com/gernest/vince/pkg/gen/client/vince/informers/externalversions"
	vince_listers "github.com/gernest/vince/pkg/gen/client/vince/listers/vince/v1alpha1"
	"github.com/gernest/vince/pkg/k8s"
	"github.com/rs/zerolog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Options struct {
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
	log    *zerolog.Logger
	top    Topology
}

func New(log *zerolog.Logger, clients k8s.Client, o Options, ready func()) *Control {
	x := Control{
		ready: ready,
		log:   log,
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
		Handler:    &enqueueWorkHandler{logger: log, queue: x.queue},
	}
	x.list.site = x.form.vince.Staples().V1alpha1().Sites().Lister()
	x.list.vince = x.form.vince.Staples().V1alpha1().Vinces().Lister()
	x.form.vince.Staples().V1alpha1().Sites().Informer().AddEventHandler(handler)
	x.form.vince.Staples().V1alpha1().Vinces().Informer().AddEventHandler(handler)
	x.top = Topology{
		vinceLister:   x.form.vince.Staples().V1alpha1().Vinces().Lister(),
		siteLister:    x.form.vince.Staples().V1alpha1().Sites().Lister(),
		podLister:     x.form.k8s.Core().V1().Pods().Lister(),
		secretsLister: x.form.k8s.Core().V1().Secrets().Lister(),
	}
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
	c.log.Debug().Msg("Initializing vince controller")
	// we only watch Site resources
	{
		timeout, _ := context.WithTimeout(context.Background(), 10*time.Second)
		c.log.Debug().Msg("Starting sites informer")
		c.form.vince.Start(ctx.Done())
		for _, ok := range c.form.vince.WaitForCacheSync(timeout.Done()) {
			if !ok {
				c.log.Fatal().Msg("timed out waiting for controller caches to sync:")
			}
		}
	}
	c.ready()
	c.log.Debug().Msg("Controller is ready")
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
