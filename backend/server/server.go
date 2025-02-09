package server

import (
	"backend/configs"
	"net/http"
	"time"
)

type Server struct {
	server *http.Server
}

func NewServer(cfg *configs.ServerConfig, handler http.Handler) *Server {
	return &Server{
		server: &http.Server{
			Addr:           ":" + cfg.Port,
			Handler:        handler,
			ReadTimeout:    cfg.ReadTimeout * time.Second,
			WriteTimeout:   cfg.WriteTimeout * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
	}
}

func (s *Server) Run() error {
	return s.server.ListenAndServe()
}
