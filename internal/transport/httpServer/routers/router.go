package routers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/krassor/GoCV/internal/transport/httpServer/handlers"
)

type FrRouter struct {
	DHandler handlers.DeviceHandlers
}

func NewDnnTrainerRouter(deviceHandler handlers.DeviceHandlers) *FrRouter {
	return &FrRouter{
		FrHandler: frHandler,
	}
}

func (frr *FrRouter) Router(r *chi.Mux) {
	r.Use(cors.AllowAll().Handler)
	r.Post("/devices", frr.FrHandler.Train)
}
