package vince

import (
	"net/http"
	"time"

	"github.com/gernest/vince/render"
)

type Health struct {
	DB      bool         `json:"sqlite"`
	Workers WorkerHealth `json:"workers"`
}

func (v *Vince) health(w http.ResponseWriter, r *http.Request) {
	var sqliteHealth bool
	if db, err := v.sql.DB(); err == nil {
		if _, err := db.Query("SELECT 1"); err == nil {
			sqliteHealth = true
		}
	}
	h := Health{
		DB:      sqliteHealth,
		Workers: v.hs.Check(),
	}
	code := http.StatusOK
	if !h.DB || !h.Workers.EventWriter || !h.Workers.SessionWriter || !h.Workers.SeriesFlush {
		code = http.StatusInternalServerError
	}
	render.JSON(w, code, h)
}

type WorkerHealthChannels struct {
	eventWriter   chan chan struct{}
	sessionWriter chan chan struct{}
	seriesFlush   chan chan struct{}
}

func newWorkerHealth() *WorkerHealthChannels {
	return &WorkerHealthChannels{
		eventWriter:   make(chan chan struct{}, 1),
		sessionWriter: make(chan chan struct{}, 1),
		seriesFlush:   make(chan chan struct{}, 1),
	}
}

func (w *WorkerHealthChannels) ping(ch chan chan struct{}) bool {
	ts := time.NewTimer(time.Millisecond)
	defer ts.Stop()
	s := make(chan struct{}, 1)
	defer close(s)

	select {
	case ch <- s:
	case <-ts.C:
		return false
	}
	ts.Reset(time.Millisecond)
	select {
	case <-s:
	case <-ts.C:
		return false
	}
	return true
}

func (w *WorkerHealthChannels) Check() WorkerHealth {
	return WorkerHealth{
		EventWriter:   w.ping(w.eventWriter),
		SessionWriter: w.ping(w.sessionWriter),
		SeriesFlush:   w.ping(w.seriesFlush),
	}
}

func (w *WorkerHealthChannels) Close() {
	close(w.eventWriter)
	close(w.sessionWriter)
	close(w.seriesFlush)
}

type WorkerHealth struct {
	EventWriter   bool `json:"event_writer"`
	SessionWriter bool `json:"session_writer"`
	SeriesFlush   bool `json:"series_flush"`
}
