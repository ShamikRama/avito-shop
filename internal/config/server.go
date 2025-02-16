package config

import (
	"net/http"
	"time"
)

type HttpServer struct {
	server http.Server
}

func InitHttpServer(cfg Config, router http.Handler) *HttpServer {
	return &HttpServer{
		server: http.Server{
			Addr:         cfg.Address,
			Handler:      router,
			ReadTimeout:  cfg.HTTPServer.Timeout * time.Second,
			WriteTimeout: cfg.HTTPServer.Timeout * time.Second,
			IdleTimeout:  cfg.HTTPServer.IdleTimeout * time.Second,
		},
	}
}

func RunServer(server *HttpServer) error {
	if err := server.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
