package server

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/caddyserver/certmagic"
	"github.com/vinceanalytics/vince/assets"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/health"
	"github.com/vinceanalytics/vince/internal/log"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/plug"
	"github.com/vinceanalytics/vince/internal/router"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/internal/worker"
	"golang.org/x/sync/errgroup"
)

func Serve(o *config.Options) error {
	ctx, err := config.Load(o)
	if err != nil {
		return err
	}
	ctx, resources := Configure(ctx, o)
	if err != nil {
		return err
	}
	return Run(ctx, resources)
}

type ResourceList []io.Closer

type resourceFunc func() error

func (r resourceFunc) Close() error {
	return r()
}

func (r ResourceList) Close() error {
	e := make([]error, 0, len(r))
	for i := len(r) - 1; i > 0; i-- {
		e = append(e, r[i].Close())
	}
	return errors.Join(e...)
}

func Configure(ctx context.Context, o *config.Options) (context.Context, ResourceList) {
	log.Level(o.GetLogLevel())
	o.Validate()

	var resources ResourceList

	// we start listeners early to make sure we can actually bind to the network.
	// This saves us managing all long running goroutines we start in this process.
	httpListener := must.Must(net.Listen("tcp", o.Listen))(
		"failed binding network address", o.Listen,
	)
	resources = append(resources, httpListener)
	ctx = core.SetHTTPListener(ctx, httpListener)
	var httpsListener net.Listener
	var magic *certmagic.Config
	if o.TLS.Enabled {
		if o.Acme.Enabled {
			magic = o.Magic()
			must.One(magic.ManageSync(ctx, []string{o.Acme.Domain}))(
				"failed to sync acme domain",
			)
			httpsListener = must.Must(net.Listen("tcp", o.TLS.Address))(
				"failed to bind https socket", o.TLS.Address,
			)
		} else {
			cert := must.Must(tls.LoadX509KeyPair(o.TLS.Cert, o.TLS.Key))(
				"failed to load tls certificates",
			)
			config := tls.Config{}
			config.Certificates = append(config.Certificates, cert)
			httpsListener = must.Must(tls.Listen("tcp", o.TLS.Address, &config))(
				"failed to bind tls socket with tls config",
			)
		}
	}
	if httpsListener != nil {
		resources = append(resources, httpsListener)
		ctx = core.SetHTTPSListener(ctx, httpsListener)
	}
	ctx, dba := db.Open(ctx, o.DataPath)
	resources = append(resources, dba)
	ctx, ts := timeseries.Open(ctx, o)
	resources = append(resources, ts)

	h := &health.Config{}
	resources = append(resources, h)
	ctx = health.Set(ctx, h)

	// configure http server
	httpSvr := &http.Server{
		Handler:           Handle(ctx),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       5 * time.Second,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}
	if httpsListener != nil {
		httpSvr.Handler = redirect(httpsListener.Addr().String())
		if magic != nil {
			// We are using tls with auto tls
			httpSvr.Handler = magic.Issuers[0].(*certmagic.ACMEIssuer).HTTPChallengeHandler(
				redirect(httpsListener.Addr().String()),
			)
		}
	}
	ctx = core.SetHTTPServer(ctx, httpSvr)
	resources = append(resources, httpSvr)

	if httpsListener != nil {
		//configure https server
		httpsSvr := &http.Server{
			Handler:           Handle(ctx),
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      2 * time.Minute,
			IdleTimeout:       5 * time.Minute,
			BaseContext: func(l net.Listener) context.Context {
				return ctx
			},
		}
		if magic != nil {
			// httpsListener is not wrapped with tls yet. We use certmagic to obtain
			// tls Config and properly wrap it.
			tlsConfig := magic.TLSConfig()
			tlsConfig.NextProtos = append([]string{"h2", "http/1.1"}, tlsConfig.NextProtos...)
			httpsListener = tls.NewListener(httpsListener, tlsConfig)
			ctx = core.SetHTTPSListener(ctx, httpsListener)
		}
		ctx = core.SetHTTPSServer(ctx, httpsSvr)
		resources = append(resources, httpsSvr)

	}
	return ctx, resources
}

func Run(ctx context.Context, resources ResourceList) error {
	var cancel context.CancelFunc
	if config.Get(ctx).NoSignal {
		ctx, cancel = context.WithCancel(ctx)
	} else {
		ctx, cancel = signal.NotifyContext(ctx, os.Interrupt)
	}
	defer cancel()

	h := health.Get(ctx)
	var g errgroup.Group
	{
		o := config.Get(ctx)
		g.Go(worker.Periodic(ctx, h.Ping("series"), o.Intervals.TSSync, false, worker.SaveBuffers))
	}

	plain := core.GetHTTPServer(ctx)
	secure := core.GetHTTPSServer(ctx)
	plainLS := core.GetHTTPListener(ctx)
	secureLS := core.GetHTTPSListener(ctx)

	g.Go(func() error {
		return plain.Serve(plainLS)
	})
	if secure != nil {
		g.Go(func() error {
			return secure.Serve(secureLS)
		})
	}
	g.Go(func() error {
		// Ensure we close the servers.
		<-ctx.Done()
		log.Get().Debug().Msg("shutting down gracefully ")
		return resources.Close()
	})
	log.Get().Debug().Str("address", plainLS.Addr().String()).Msg("started serving  http traffic")
	if secureLS != nil {
		log.Get().Debug().Str("address", secureLS.Addr().String()).Msg("started serving  https traffic")
	}
	return g.Wait()
}

func Handle(ctx context.Context) http.Handler {
	pipe := append(
		plug.Pipeline{
			plug.Track(),
			assets.Plug(),
			plug.RequestID,
			plug.CORS,
		},
		router.Pipe(ctx)...,
	)
	h := pipe.Pass(plug.NOOP)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

func redirect(addr string) http.Handler {
	_, port, _ := net.SplitHostPort(addr)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		toURL := "https://"
		requestHost := hostOnly(r.Host)
		toURL += requestHost + ":" + port
		toURL += r.URL.RequestURI()
		w.Header().Set("Connection", "close")
		http.Redirect(w, r, toURL, http.StatusMovedPermanently)
	})
}

// hostOnly returns only the host portion of hostport.
// If there is no port or if there is an error splitting
// the port off, the whole input string is returned.
func hostOnly(hostPort string) string {
	host, _, err := net.SplitHostPort(hostPort)
	if err != nil {
		return hostPort // OK; probably had no port to begin with
	}
	return host
}
