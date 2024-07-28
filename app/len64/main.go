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
	mux.HandleFunc("/", db.Wrap(
		plug.BrowserHome().Then(web.Home),
	))
	mux.HandleFunc("GET /login", db.Wrap(
		plug.BrowserFormGet().Then(web.LoginForm),
	))
	mux.HandleFunc("POST /login", db.Wrap(
		plug.BrowserFormPost().Then(web.Login),
	))
	mux.HandleFunc("GET /register", db.Wrap(
		plug.BrowserFormGet().Then(web.RegisterForm),
	))
	mux.HandleFunc("POST /register", db.Wrap(
		plug.BrowserFormPost().Then(web.Register),
	))

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
