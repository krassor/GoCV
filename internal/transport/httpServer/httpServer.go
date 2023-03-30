package httpServer

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/krassor/GoCV/internal/transport/httpServer/routers"
	"github.com/rs/zerolog/log"
)

type HttpServer struct {
	Router     *routers.FrRouter
	httpServer *http.Server
}

func NewHttpServer(router *routers.FrRouter) *HttpServer {
	return &HttpServer{
		Router: router,
	}
}

func (h *HttpServer) Listen() {
	app := chi.NewRouter()
	h.Router.Router(app)

	serverPort, ok := os.LookupEnv("FR_HTTP_PORT")
	if !ok {
		serverPort = "8080"
	}
	serverAddress, ok := os.LookupEnv("FR_HTTP_HOST_LISTEN")
	if !ok {
		serverAddress = "127.0.0.1"
	}
	log.Info().Msgf("Server http get env %s:%s ", serverAddress, serverPort)

	h.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%s", serverAddress, serverPort),
		Handler: app,
	}
	log.Info().Msgf("Server started on Port %s ", serverPort)

	err := h.httpServer.ListenAndServe()

	if err != nil && err != http.ErrServerClosed {
		log.Warn().Msgf("httpServer.ListenAndServe() Error: %s", err)
	}

	if err == http.ErrServerClosed {
		log.Info().Msgf("%s", err)
	}

}

func (h *HttpServer) Shutdown(ctx context.Context) error {
	if err := h.httpServer.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}
