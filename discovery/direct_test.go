package discovery

import (
	"context"
	"testing"
)

func TestDirectRegistryResolve(t *testing.T) {
	reg := newDirectRegistry(map[string]string{"sys_roles": "127.0.0.1:9001"})
	addrs, err := reg.Resolve(context.Background(), "sys_roles")
	if err != nil || len(addrs) != 1 || addrs[0] != "127.0.0.1:9001" {
		t.Fatalf("resolve: %v %v", addrs, err)
	}
}
