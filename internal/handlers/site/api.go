package site

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
)

func APIGet(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, http.StatusOK, models.GetSite(r.Context()))
}

func APIUpdate(w http.ResponseWriter, r *http.Request) {
	render.JSONError(w, http.StatusNotImplemented, http.StatusText(http.StatusNotImplemented))
}

func APIList(w http.ResponseWriter, r *http.Request) {
	render.JSONError(w, http.StatusNotImplemented, http.StatusText(http.StatusNotImplemented))
}

func APIDelete(w http.ResponseWriter, r *http.Request) {
	render.JSONError(w, http.StatusNotImplemented, http.StatusText(http.StatusNotImplemented))
}

func APICreate(w http.ResponseWriter, r *http.Request) {
	render.JSONError(w, http.StatusNotImplemented, http.StatusText(http.StatusNotImplemented))
}
