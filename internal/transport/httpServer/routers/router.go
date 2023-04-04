package routers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/krassor/GoCV/internal/transport/httpServer/handlers"
)

type FrRouter struct {
	FrHandler *handlers.FrHandler
}

func NewDnnTrainerRouter(frHandler *handlers.FrHandler) *FrRouter {
	return &FrRouter{
		FrHandler: frHandler,
	}
}

func (frr *FrRouter) Router(r *chi.Mux) {
	r.Use(cors.AllowAll().Handler)
	r.Post("/tenant", frr.FrHandler.LoadTenantFoto)
	//r.Get("/debug/pprof/*", http.DefaultServeMux.ServeHTTP)

}
