package server

import (
	"net/http"
)

type Server struct {
	Addr   string
	Handler http.Handler
}

func New(addr string, handler http.Handler) *Server {
	return &Server{
		Addr:    addr,
		Handler: handler,
	}
}

func (s *Server) Run() error {
	return http.ListenAndServe(s.Addr, s.Handler)
}
