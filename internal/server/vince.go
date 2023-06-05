package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/caddyserver/certmagic"
	"github.com/vinceanalytics/vince/assets"
	"github.com/vinceanalytics/vince/internal/alerts"
	"github.com/vinceanalytics/vince/internal/caches"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/email"
	"github.com/vinceanalytics/vince/internal/health"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/plug"
	"github.com/vinceanalytics/vince/internal/router"
	"github.com/vinceanalytics/vince/internal/sessions"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/internal/userid"
	"github.com/vinceanalytics/vince/internal/worker"
	"github.com/vinceanalytics/vince/pkg/log"
	"golang.org/x/sync/errgroup"
)

func Serve(o *config.Options) error {
	ctx, err := config.Load(o)
	if err != nil {
		return err
	}
	ctx, resources, err := Configure(ctx, o)
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
	e := make([]error, len(r))
	for i, f := range r {
		e[i] = f.Close()
	}
	return errors.Join(e...)
}

func Configure(ctx context.Context, o *config.Options) (context.Context, ResourceList, error) {
	var resources ResourceList

	// we start listeners early to make sure we can actually bind to the network.
	// This saves us managing all long running goroutines we start in this process.
	httpListener, err := net.Listen("tcp", o.Listen)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to bind to a network address %v", err)
	}
	resources = append(resources, httpListener)
	ctx = core.SetHTTPListener(ctx, httpListener)
	var httpsListener net.Listener
	var magic *certmagic.Config
	if o.TLS.Enabled {
		if o.TLS.Address == "" {
			resources.Close()
			return nil, nil, errors.New("tls-address is required")
		}
		if o.TLS.Key == "" || o.TLS.Cert == "" {
			if !o.Acme.Enabled {
				resources.Close()
				return nil, nil, errors.New("tls-key and tls-cert  are required")
			}
		}
		if o.Acme.Enabled {
			if o.Acme.Email == "" || o.Acme.Domain == "" {
				resources.Close()
				return nil, nil, errors.New("acme-email and acme-domain  are required")
			}
			magic = certmagic.NewDefault()
			// we use file storage for certs
			certsPath := filepath.Join(o.DataPath, "certs")
			os.MkdirAll(certsPath, 0755)
			magic.Storage = &certmagic.FileStorage{Path: certsPath}
			myACME := certmagic.NewACMEIssuer(magic, certmagic.ACMEIssuer{
				CA:     certmagic.LetsEncryptStagingCA,
				Email:  o.Acme.Email,
				Agreed: true,
			})
			magic.Issuers = append(magic.Issuers, myACME)
			err = magic.ManageSync(ctx, []string{o.Acme.Domain})
			if err != nil {
				resources.Close()
				return nil, nil, fmt.Errorf("failed to sync acme domain %v", err)
			}
			httpsListener, err = net.Listen("tcp", o.TLS.Address)
			if err != nil {
				resources.Close()
				return nil, nil, fmt.Errorf("failed to bind to https socket %v", err)
			}
		} else {
			cert, err := tls.LoadX509KeyPair(o.TLS.Cert, o.TLS.Key)
			if err != nil {
				resources.Close()
				return nil, nil, fmt.Errorf("failed to load https certificate %v", err)
			}
			config := tls.Config{}
			config.Certificates = append(config.Certificates, cert)
			httpsListener, err = tls.Listen("tcp", o.TLS.Address, &config)
			if err != nil {
				resources.Close()
				return nil, nil, fmt.Errorf("failed to bind https socket %v", err)
			}
		}
	}
	if httpsListener != nil {
		resources = append(resources, httpsListener)
		ctx = core.SetHTTPSListener(ctx, httpsListener)
	}

	ctx = userid.Open(ctx)

	sqlDb, err := models.Open(models.Database(o))
	if err != nil {
		resources.Close()
		return nil, nil, err
	}
	resources = append(resources, resourceFunc(func() error {
		return models.CloseDB(sqlDb)
	}))

	ctx = models.Set(ctx, sqlDb)

	if o.Bootstrap.Enabled {
		log.Get().Debug().Msg("bootstrapping user")
		if o.Bootstrap.Name == "" ||
			o.Bootstrap.Email == "" ||
			o.Bootstrap.Password == "" ||
			o.Bootstrap.Key == "" {
			resources.Close()
			return nil, nil, errors.New("bootstrap-name, bootstrap-email, bootstrap-password, and bootstrap-key, are required")
		}
		models.Bootstrap(ctx,
			o.Bootstrap.Name, o.Bootstrap.Email, o.Bootstrap.Password, o.Bootstrap.Key,
		)
	}
	if o.Alerts.Enabled {
		log.Get().Debug().Msg("setup alerts")
		a, err := alerts.Setup(o)
		if err != nil {
			resources.Close()
			return nil, nil, err
		}
		ctx = alerts.Set(ctx, a)
	}
	if o.Mailer.Enabled {
		log.Get().Debug().Msg("setup mailer")
		mailer, err := email.FromConfig(o)
		if err != nil {
			log.Get().Err(err).Msg("failed creating mailer")
			resources.Close()
			return nil, nil, err
		}
		resources = append(resources, mailer)
		ctx = email.Set(ctx, mailer)
	}
	ctx, ts, err := timeseries.Open(ctx, o)
	if err != nil {
		resources.Close()
		return nil, nil, err
	}
	resources = append(resources, ts)
	ctx, err = caches.Open(ctx)
	if err != nil {
		log.Get().Err(err).Msg("failed to open caches")
		resources.Close()
		return nil, nil, err
	}
	resources = append(resources, resourceFunc(func() error {
		return caches.Close(ctx)
	}))

	session := sessions.NewSession("_vince")
	ctx = sessions.Set(ctx, session)
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
	svr := ResourceList{httpSvr}

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
		svr = append(svr, httpsSvr)

	}
	resources = append(svr, resources...)
	return ctx, resources, nil
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
	addHealth := func(x *health.Ping) {
		h.Health = append(h.Health, x)
	}
	h.Health = append(h.Health, health.Base{
		Key:       "database",
		CheckFunc: models.Check,
	})
	{
		// register and start workers
		g.Go(worker.UpdateCacheSites(ctx, addHealth))
		g.Go(worker.SaveTimeseries(ctx, addHealth))
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
			plug.Favicon(plug.DefaultClient),
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
func hostOnly(hostport string) string {
	host, _, err := net.SplitHostPort(hostport)
	if err != nil {
		return hostport // OK; probably had no port to begin with
	}
	return host
}
