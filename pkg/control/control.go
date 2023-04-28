package control

import (
	"context"
	"fmt"
	"sync"
	"time"

	siteinformer "github.com/gernest/vince/pkg/gen/client/site/informers/externalversions"
	sitelisterr "github.com/gernest/vince/pkg/gen/client/site/listers/site/v1alpha1"
	"github.com/gernest/vince/pkg/k8s"
	"github.com/gernest/vince/pkg/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	// configRefreshKey is the work queue key used to indicate that config has to be refreshed.
	configRefreshKey = "refresh"

	// maxRetries is the number of times a work task will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the times a
	// work task is going to be re-queued: 5ms, 10ms, 20ms, 40ms, 80ms, 160ms, 320ms, 640ms, 1.3s, 2.6s, 5.1s, 10.2s.
	maxRetries = 12
)

type Options struct {
	Namespace        string
	WatchNamespaces  []string
	IgnoreNamespaces []string
}

type Control struct {
	mu      sync.Mutex
	stop    chan struct{}
	opts    Options
	work    workqueue.RateLimitingInterface
	clients k8s.Client
	form    Inform
	list    List
	filter  *k8s.ResourceFilter
}

func New(ctx context.Context, clients k8s.Client, o Options) *Control {
	x := Control{
		stop: make(chan struct{}),
		opts: o,
		filter: k8s.NewResourceFilter(
			k8s.WatchNamespaces(o.WatchNamespaces...),
			k8s.IgnoreNamespaces(o.IgnoreNamespaces...),
			k8s.IgnoreNamespaces(metav1.NamespaceSystem),
		),
		form: Inform{
			site: siteinformer.NewSharedInformerFactory(clients.Site(), k8s.ResyncPeriod),
		},
		work: workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
	}
	handler := cache.FilteringResourceEventHandler{
		FilterFunc: x.isWatchedResource,
		Handler:    &enqueueWorkHandler{logger: log.Get(ctx), workQueue: x.work},
	}
	x.list.site = x.form.site.Vince().V1alpha1().Sites().Lister()
	x.form.site.Vince().V1alpha1().Sites().Informer().AddEventHandler(handler)
	return &x
}

type Inform struct {
	site siteinformer.SharedInformerFactory
}

type List struct {
	site sitelisterr.SiteLister
}

func (c *Control) Run(ctx context.Context) error {
	// Handle a panic with logging and exiting.
	defer utilruntime.HandleCrash()
	waitGroup := sync.WaitGroup{}
	defer func() {
		log.Get(ctx).Info().Msg("Shutting down workers")
		c.work.ShutDown()
		waitGroup.Wait()
	}()
	x := log.Get(ctx)
	x.Debug().Msg("Initializing vince controller")
	err := c.startInformers(ctx, 10*time.Second)
	if err != nil {
		return err
	}
	// Start to poll work from the queue.
	waitGroup.Add(1)

	runWorker := func() {
		defer waitGroup.Done()
		c.runWorker()
	}
	go wait.Until(runWorker, time.Second, c.stop)
	<-c.stop
	return nil
}

// isWatchedResource returns true if the given resource is not ignored, false otherwise.
func (c *Control) isWatchedResource(obj interface{}) bool {
	return !c.filter.IsIgnored(obj)
}

func (c *Control) startInformers(ctx context.Context, timeout time.Duration) error {
	ctx, _ = context.WithTimeout(ctx, timeout)
	log.Get(ctx).Debug().Msg("Starting Informers")
	c.form.site.Start(c.stop)
	for t, ok := range c.form.site.WaitForCacheSync(ctx.Done()) {
		if !ok {
			return fmt.Errorf("timed out waiting for controller caches to sync: %s", t)
		}
	}
	return nil

}

func (c *Control) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *Control) processNextWorkItem() bool {
	key, ok := c.work.Get()
	if !ok {
		return ok
	}
	defer c.work.Done(key)
	c.work.Forget(key)
	return true
}
