package grpcx

import (
	"errors"

	"github.com/txbao/goeasy/config"
)

// Server gRPC 服务占位。
type Server struct {
	addr string
}

func NewServer(cfg config.GRPC) (*Server, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	if cfg.Addr == "" {
		return nil, errors.New("grpc enabled but addr is empty")
	}
	return &Server{addr: cfg.Addr}, nil
}

func (s *Server) Addr() string {
	if s == nil {
		return ""
	}
	return s.addr
}
