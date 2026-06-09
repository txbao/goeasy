package grpcx

import (
	"testing"

	"github.com/txbao/goeasy/config"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

func TestNewServerRegistersReflection(t *testing.T) {
	s, err := NewServer(config.GRPC{Enabled: true, Addr: "127.0.0.1:0"})
	if err != nil {
		t.Fatal(err)
	}
	if s == nil || s.GRPC() == nil {
		t.Fatal("expected grpc server")
	}
	info := s.GRPC().GetServiceInfo()
	if _, ok := info[grpc_reflection_v1alpha.ServerReflection_ServiceDesc.ServiceName]; !ok {
		t.Fatalf("reflection service not registered, services: %v", info)
	}
	_ = s.Shutdown(t.Context())
}
