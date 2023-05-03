package control

import (
	"context"
	"time"

	vince "github.com/gernest/vince/pkg/apis/vince/v1alpha1"
	vince_informers "github.com/gernest/vince/pkg/gen/client/vince/informers/externalversions"
	vince_listers "github.com/gernest/vince/pkg/gen/client/vince/listers/vince/v1alpha1"
	"github.com/gernest/vince/pkg/k8s"
	"github.com/rs/zerolog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
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
			vince: vince_informers.NewSharedInformerFactory(clients.Vince(), k8s.ResyncPeriod),
		},
		work: make(chan *Work, 2<<10),
	}
	handler := cache.FilteringResourceEventHandler{
		FilterFunc: x.isWatchedResource,
		Handler:    &enqueueWorkHandler{logger: log, workQueue: x.work},
	}
	x.list.site = x.form.vince.Vince().V1alpha1().Sites().Lister()
	x.list.vince = x.form.vince.Vince().V1alpha1().Vinces().Lister()
	x.form.vince.Vince().V1alpha1().Sites().Informer().AddEventHandler(handler)
	x.form.vince.Vince().V1alpha1().Vinces().Informer().AddEventHandler(handler)
	return &x
}

type Inform struct {
	vince vince_informers.SharedInformerFactory
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
	for {
		select {
		case <-ctx.Done():
			c.log.Debug().Msg("shutting down the controller")
			return ctx.Err()
		case w := <-c.work:
			switch e := w.Item.(type) {
			case *vince.Site:
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
