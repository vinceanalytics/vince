package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/gernest/len64/app"
	"github.com/gernest/len64/web"
	"github.com/gernest/len64/web/db"
	"github.com/gernest/len64/web/db/plug"
)

func main() {
	dataPath := flag.String("data", ".data", "Path to where database data is stored")
	flag.Parse()
	db, err := db.Open(*dataPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	err = db.Start(ctx)
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.Handle("/public/", plug.Track(http.FileServerFS(app.Public)))

	mux.HandleFunc("/favicon/sources/placeholder", web.Placeholder)
	mux.HandleFunc("/favicon/sources/{source...}", web.Favicon)

	mux.HandleFunc("/{$}", db.Wrap(
		plug.Browser().Then(web.Home),
	))

	mux.HandleFunc("GET /login", db.Wrap(
		plug.Browser().
			With(plug.CSRF).
			Then(web.LoginForm),
	))

	mux.HandleFunc("POST /login", db.Wrap(
		plug.Browser().
			With(plug.VerifyCSRF).
			With(plug.RequireLogout).
			Then(web.Login),
	))

	mux.HandleFunc("GET /register", db.Wrap(
		plug.Browser().
			With(plug.CSRF).
			With(plug.Captcha).
			Then(web.RegisterForm),
	))

	mux.HandleFunc("POST /register", db.Wrap(
		plug.Browser().
			With(plug.VerifyCSRF).
			Then(web.Register),
	))

	mux.HandleFunc("GET /sites/new", db.Wrap(
		plug.Browser().
			With(plug.CSRF).
			With(plug.RequireAccount).
			Then(web.CreateSiteForm),
	))

	mux.HandleFunc("POST /sites", db.Wrap(
		plug.Browser().
			With(plug.CSRF).
			With(plug.RequireAccount).
			Then(web.CreateSite),
	))

	mux.HandleFunc("GET /sites", db.Wrap(
		plug.Browser().
			With(plug.RequireAccount).
			Then(web.Sites),
	))

	mux.HandleFunc("GET /avatar/{size}/{uid...}", db.Wrap(
		plug.Browser().
			With(plug.RequireAccount).
			Then(web.Avatar),
	))

	mux.HandleFunc("/api/event", db.Wrap(web.Event))

	svr := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	slog.Info("starting server", "addr", svr.Addr)
	go func() {
		defer cancel()
		svr.ListenAndServe()
	}()
	<-ctx.Done()
	svr.Close()
	slog.Info("Shutting down")
}
