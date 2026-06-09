package discovery

import (
	"testing"

	"github.com/txbao/goeasy/config"
)

func TestResolveAdvertiseAddr(t *testing.T) {
	disc := config.Discovery{
		Etcd: config.EtcdDiscoveryCfg{AdvertiseAddr: "10.0.0.1:9001"},
		Services: map[string]string{
			"demo": "10.0.0.2:9001",
		},
	}
	if got := ResolveAdvertiseAddr("demo", "0.0.0.0:9001", disc); got != "10.0.0.1:9001" {
		t.Fatalf("advertise_addr: got %q", got)
	}

	disc.Etcd.AdvertiseAddr = ""
	if got := ResolveAdvertiseAddr("demo", "0.0.0.0:9001", disc); got != "10.0.0.2:9001" {
		t.Fatalf("services map: got %q", got)
	}

	if got := ResolveAdvertiseAddr("other", "0.0.0.0:9001", disc); got != "127.0.0.1:9001" {
		t.Fatalf("normalize 0.0.0.0: got %q", got)
	}

	if got := ResolveAdvertiseAddr("other", "192.168.1.5:28021", disc); got != "192.168.1.5:28021" {
		t.Fatalf("keep listen addr: got %q", got)
	}
}
