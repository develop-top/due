package direct

import (
	"context"
	"github.com/develop-top/due/v2/cluster"
	"github.com/develop-top/due/v2/core/endpoint"
	"github.com/develop-top/due/v2/errors"
	"github.com/develop-top/due/v2/log"
	"github.com/develop-top/due/v2/registry"
	"google.golang.org/grpc/resolver"
	"net"
	"sync"
	"time"
)

const scheme = "direct"

const defaultTimeout = 10 * time.Second

type Builder struct {
	dis       registry.Discovery
	ctx       context.Context
	cancel    context.CancelFunc
	watcher   registry.Watcher
	rw        sync.RWMutex
	addresses map[string]string
}

var _ resolver.Builder = &Builder{}

func NewBuilder(dis registry.Discovery) *Builder {
	b := &Builder{}
	b.dis = dis
	b.ctx, b.cancel = context.WithCancel(context.Background())
	b.addresses = make(map[string]string)

	if err := b.init(); err != nil {
		log.Fatalf("init client builder failed: %v", err)
	}

	return b
}

func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	addr := target.URL.Host

	if _, _, err := net.SplitHostPort(target.URL.Host); err != nil {
		b.rw.RLock()
		address, ok := b.addresses[target.URL.Host]
		b.rw.RUnlock()
		if !ok {
			return nil, errors.ErrNotFoundDirectAddress
		}

		addr = address
	}

	if err := cc.UpdateState(resolver.State{Addresses: []resolver.Address{{Addr: addr}}}); err != nil {
		return nil, err
	}

	return newResolver(), nil
}

func (b *Builder) Scheme() string {
	return scheme
}

func (b *Builder) init() error {
	if b.dis == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(b.ctx, defaultTimeout)
	instances, err := b.dis.Services(ctx, cluster.Mesh.String())
	cancel()
	if err != nil {
		return err
	}

	ctx, cancel = context.WithTimeout(b.ctx, defaultTimeout)
	watcher, err := b.dis.Watch(ctx, cluster.Mesh.String())
	cancel()
	if err != nil {
		return err
	}

	b.watcher = watcher
	b.updateInstances(instances)

	go b.watch()

	return nil
}

func (b *Builder) watch() {
	for {
		select {
		case <-b.ctx.Done():
			return
		default:
			// exec watch
		}
		instances, err := b.watcher.Next()
		if err != nil {
			continue
		}

		b.updateInstances(instances)
	}
}

func (b *Builder) updateInstances(instances []*registry.ServiceInstance) {
	addresses := make(map[string]string, len(instances))
	for _, instance := range instances {
		ep, err := endpoint.ParseEndpoint(instance.Endpoint)
		if err != nil {
			log.Errorf("parse discovery endpoint failed: %v", err)
			continue
		}

		addresses[instance.ID] = ep.Address()
	}

	b.rw.Lock()
	b.addresses = addresses
	b.rw.Unlock()
}
