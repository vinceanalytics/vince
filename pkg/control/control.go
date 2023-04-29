package control

import (
	"context"
	"time"

	"github.com/gernest/vince/pkg/apis/site/v1alpha1"
	siteinformer "github.com/gernest/vince/pkg/gen/client/site/informers/externalversions"
	sitelisterr "github.com/gernest/vince/pkg/gen/client/site/listers/site/v1alpha1"
	"github.com/gernest/vince/pkg/k8s"
	"github.com/rs/zerolog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
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
	opts   Options
	work   chan *Work
	form   Inform
	list   List
	filter *k8s.ResourceFilter
	ready  func()
	log    *zerolog.Logger
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
			site: siteinformer.NewSharedInformerFactory(clients.Site(), k8s.ResyncPeriod),
		},
		work: make(chan *Work, 2<<10),
	}
	handler := cache.FilteringResourceEventHandler{
		FilterFunc: x.isWatchedResource,
		Handler:    &enqueueWorkHandler{logger: log, workQueue: x.work},
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
	c.log.Debug().Msg("Initializing vince controller")
	// we only watch Site resources
	{
		timeout, _ := context.WithTimeout(context.Background(), 10*time.Second)
		c.log.Debug().Msg("Starting sites informer")
		c.form.site.Start(ctx.Done())
		for _, ok := range c.form.site.WaitForCacheSync(timeout.Done()) {
			if !ok {
				c.log.Fatal().Msg("timed out waiting for controller caches to sync:")
			}
		}
	}
	c.ready()
	c.log.Debug().Msg("Controller is ready")
	for {
		select {
		case <-ctx.Done():
			c.log.Debug().Msg("shutting down the controller")
			return ctx.Err()
		case w := <-c.work:
			switch e := w.Item.(type) {
			case *v1alpha1.Site:
				switch w.Op {
				case ADD:
					c.log.Debug().
						Str("name", e.Name).
						Str("ns", e.Namespace).
						Msg("adding site")
				case Update:
					c.log.Debug().
						Str("name", e.Name).
						Str("ns", e.Namespace).
						Msg("updating site")
				case Delete:
					c.log.Debug().
						Str("name", e.Name).
						Str("ns", e.Namespace).
						Msg("deleting site")
				}
			}
		}
	}
}

// isWatchedResource returns true if the given resource is not ignored, false otherwise.
func (c *Control) isWatchedResource(obj interface{}) bool {
	return !c.filter.IsIgnored(obj)
}
