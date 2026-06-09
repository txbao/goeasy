package grpcx

import (
	"context"
	"errors"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/txbao/goeasy/config"
)

// Server gRPC 服务端封装（Listen 延迟到 Serve，避免 InitInfra 占用端口）。
type Server struct {
	cfg config.GRPC
	gs  *grpc.Server
	lis net.Listener
	mu  sync.Mutex
}

func NewServer(cfg config.GRPC) (*Server, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	if cfg.Addr == "" {
		return nil, errors.New("grpc enabled but addr is empty")
	}
	gs := grpc.NewServer()
	reflection.Register(gs)
	return &Server{cfg: cfg, gs: gs}, nil
}

func (s *Server) ensureListen() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lis != nil {
		return nil
	}
	lis, err := net.Listen("tcp", s.cfg.Addr)
	if err != nil {
		return err
	}
	s.lis = lis
	return nil
}

// GRPC 返回原生 *grpc.Server，供业务 RegisterXxxServer 使用。
func (s *Server) GRPC() *grpc.Server {
	if s == nil {
		return nil
	}
	return s.gs
}

// EnsureListen 提前绑定监听（供服务发现注册前获取实际监听地址）。
func (s *Server) EnsureListen() error {
	if s == nil {
		return nil
	}
	return s.ensureListen()
}

func (s *Server) Addr() string {
	if s == nil {
		return ""
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lis != nil {
		return s.lis.Addr().String()
	}
	return s.cfg.Addr
}

func (s *Server) Serve() error {
	if s == nil || s.gs == nil {
		return nil
	}
	if err := s.ensureListen(); err != nil {
		return err
	}
	return s.gs.Serve(s.lis)
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s == nil || s.gs == nil {
		return nil
	}
	ch := make(chan struct{})
	go func() {
		s.gs.GracefulStop()
		close(ch)
	}()
	select {
	case <-ctx.Done():
		s.gs.Stop()
		return ctx.Err()
	case <-ch:
	}
	s.mu.Lock()
	if s.lis != nil {
		_ = s.lis.Close()
		s.lis = nil
	}
	s.mu.Unlock()
	return nil
}

// Register 在原生 Server 上注册服务的便捷方法。
func (s *Server) Register(fn func(*grpc.Server)) {
	if s == nil || fn == nil {
		return
	}
	fn(s.gs)
}
