package discovery

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/txbao/goeasy/config"
)

type etcdRegistry struct {
	client      *clientv3.Client
	prefix      string
	leaseTTL    int64
	fallback    Registry
	leaseID     clientv3.LeaseID
	leaseCancel context.CancelFunc
	mu          sync.Mutex
}

func newEtcdRegistry(cfg config.EtcdDiscoveryCfg, fallback Registry) (Registry, error) {
	if len(cfg.Endpoints) == 0 {
		return nil, fmt.Errorf("discovery: etcd endpoints required")
	}
	timeout := time.Duration(cfg.DialTimeoutSec) * time.Second
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: timeout,
	})
	if err != nil {
		return nil, err
	}
	prefix := strings.TrimSuffix(cfg.Prefix, "/")
	ttl := int64(cfg.LeaseTTLSec)
	if ttl <= 0 {
		ttl = 30
	}
	return &etcdRegistry{client: cli, prefix: prefix, leaseTTL: ttl, fallback: fallback}, nil
}

func (e *etcdRegistry) key(service string) string {
	return e.prefix + "/" + service
}

func (e *etcdRegistry) Register(ctx context.Context, service, addr string) error {
	if service == "" || addr == "" {
		return fmt.Errorf("discovery: service and addr required")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.leaseCancel != nil {
		e.leaseCancel()
		e.leaseCancel = nil
	}
	lease, err := e.client.Grant(ctx, e.leaseTTL)
	if err != nil {
		return err
	}
	_, err = e.client.Put(ctx, e.key(service), addr, clientv3.WithLease(lease.ID))
	if err != nil {
		return err
	}
	kaCtx, cancel := context.WithCancel(context.Background())
	kaCh, err := e.client.KeepAlive(kaCtx, lease.ID)
	if err != nil {
		cancel()
		return err
	}
	e.leaseID = lease.ID
	e.leaseCancel = cancel
	go e.keepAliveLoop(kaCh)
	return nil
}

func (e *etcdRegistry) keepAliveLoop(ch <-chan *clientv3.LeaseKeepAliveResponse) {
	for range ch {
	}
}

func (e *etcdRegistry) Resolve(ctx context.Context, service string) ([]string, error) {
	resp, err := e.client.Get(ctx, e.key(service), clientv3.WithPrefix())
	if err != nil {
		if e.fallback != nil {
			return e.fallback.Resolve(ctx, service)
		}
		return nil, err
	}
	var addrs []string
	for _, kv := range resp.Kvs {
		if v := string(kv.Value); v != "" {
			addrs = append(addrs, v)
		}
	}
	if len(addrs) > 0 {
		return addrs, nil
	}
	if e.fallback != nil {
		return e.fallback.Resolve(ctx, service)
	}
	return nil, fmt.Errorf("discovery: no etcd endpoints for %q", service)
}

func (e *etcdRegistry) Deregister(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.leaseCancel != nil {
		e.leaseCancel()
		e.leaseCancel = nil
	}
	if e.leaseID == 0 {
		return nil
	}
	_, err := e.client.Revoke(ctx, e.leaseID)
	e.leaseID = 0
	return err
}
