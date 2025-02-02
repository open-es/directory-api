package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"directory/internal/services"
)

func New(mux http.Handler) http.Handler {
	return mux
}

func NewMuxer(middlewares chi.Middlewares, services []services.Service) (mux *chi.Mux) {
	mux = chi.NewRouter()

	for _, mw := range middlewares {
		mux.Use(mw)
	}

	for _, service := range services {
		service.RegisterRoutes(mux)
	}

	return
}
