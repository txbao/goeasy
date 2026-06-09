package discovery

import (
	"net"
	"strings"

	"github.com/txbao/goeasy/config"
)

// ResolveAdvertiseAddr 解析注册到服务发现的可达 gRPC 地址。
// 优先级：etcd.advertise_addr > discovery.services[appName] > listenAddr（0.0.0.0 回退 127.0.0.1）。
func ResolveAdvertiseAddr(appName, listenAddr string, disc config.Discovery) string {
	if adv := strings.TrimSpace(disc.Etcd.AdvertiseAddr); adv != "" {
		return adv
	}
	if adv, ok := disc.Services[appName]; ok && strings.TrimSpace(adv) != "" {
		return strings.TrimSpace(adv)
	}
	return normalizeListenAddr(listenAddr)
}

func normalizeListenAddr(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	if host == "0.0.0.0" || host == "::" || host == "" {
		return net.JoinHostPort("127.0.0.1", port)
	}
	return addr
}
