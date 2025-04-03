package mesh

import (
	"context"
	"fmt"
	"github.com/develop-top/due/v2/cluster"
	"github.com/develop-top/due/v2/component"
	"github.com/develop-top/due/v2/core/info"
	"github.com/develop-top/due/v2/log"
	"github.com/develop-top/due/v2/registry"
	"github.com/develop-top/due/v2/transport"
	"github.com/develop-top/due/v2/utils/xcall"
	"sync"
	"sync/atomic"
)

type HookHandler func(proxy *Proxy)

type Mesh struct {
	component.Base
	opts        *options
	ctx         context.Context
	cancel      context.CancelFunc
	state       atomic.Int32
	proxy       *Proxy
	transporter transport.Server
	services    []*serviceEntity
	instance    *registry.ServiceInstance
	rw          sync.RWMutex
	hooks       map[cluster.Hook][]HookHandler
}

type serviceEntity struct {
	name     string      // 服务名称;用于定位服务发现
	desc     interface{} // 服务描述(grpc为desc描述对象; rpcx为服务路径)
	provider interface{} // 服务提供者
}

func NewMesh(opts ...Option) *Mesh {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	m := &Mesh{}
	m.opts = o
	m.hooks = make(map[cluster.Hook][]HookHandler)
	m.services = make([]*serviceEntity, 0)
	m.proxy = newProxy(m)
	m.ctx, m.cancel = context.WithCancel(o.ctx)
	m.state.Store(int32(cluster.Shut))

	return m
}

// Name 组件名称
func (m *Mesh) Name() string {
	return m.opts.name
}

// Init 初始化节点
func (m *Mesh) Init() {
	if m.opts.codec == nil {
		log.Fatal("codec component is not injected")
	}

	if m.opts.registry == nil {
		log.Fatal("registry component is not injected")
	}

	if m.opts.transporter == nil {
		log.Fatal("transporter component is not injected")
	}

	m.runHookFunc(cluster.Init)
}

// Start 启动
func (m *Mesh) Start() {
	if m.state.Swap(int32(cluster.Work)) != int32(cluster.Shut) {
		return
	}

	m.startTransportServer()

	m.registerServiceInstance()

	m.proxy.watch()

	m.printInfo()

	m.runHookFunc(cluster.Start)
}

// Close 关闭
func (m *Mesh) Close() {
	if !m.state.CompareAndSwap(int32(cluster.Work), int32(cluster.Hang)) {
		if !m.state.CompareAndSwap(int32(cluster.Busy), int32(cluster.Hang)) {
			return
		}
	}

	m.refreshServiceInstance()

	m.runHookFunc(cluster.Close)
}

// Destroy 销毁
func (m *Mesh) Destroy() {
	if m.state.Swap(int32(cluster.Shut)) == int32(cluster.Shut) {
		return
	}

	m.runHookFunc(cluster.Destroy)

	m.deregisterServiceInstance()

	m.stopTransportServer()

	m.cancel()
}

// Proxy 获取节点代理
func (m *Mesh) Proxy() *Proxy {
	return m.proxy
}

// 启动传输服务器
func (m *Mesh) startTransportServer() {
	m.opts.transporter.SetDefaultDiscovery(m.opts.registry)

	transporter, err := m.opts.transporter.NewServer()
	if err != nil {
		log.Fatalf("transport server create failed: %v", err)
	}

	m.transporter = transporter

	for _, entity := range m.services {
		if err = m.transporter.RegisterService(entity.desc, entity.provider); err != nil {
			log.Fatalf("register service failed: %v", err)
		}
	}

	go func() {
		if err = m.transporter.Start(); err != nil {
			log.Fatalf("transport server start failed: %v", err)
		}
	}()
}

// 停止传输服务器
func (m *Mesh) stopTransportServer() {
	if err := m.transporter.Stop(); err != nil {
		log.Errorf("transport server stop failed: %v", err)
	}
}

// 注册服务实例
func (m *Mesh) registerServiceInstance() {
	m.instance = &registry.ServiceInstance{
		ID:       m.opts.id,
		Name:     cluster.Mesh.String(),
		Kind:     cluster.Mesh.String(),
		Alias:    m.opts.name,
		State:    m.getState().String(),
		Weight:   m.opts.weight,
		Endpoint: m.transporter.Endpoint().String(),
		Services: make([]string, 0, len(m.services)),
	}

	for _, item := range m.services {
		m.instance.Services = append(m.instance.Services, item.name)
	}

	ctx, cancel := context.WithTimeout(m.ctx, defaultTimeout)
	defer cancel()

	if err := m.opts.registry.Register(ctx, m.instance); err != nil {
		log.Fatalf("register cluster instance failed: %v", err)
	}
}

// 刷新服务实例状态
func (m *Mesh) refreshServiceInstance() {
	if m.instance == nil {
		return
	}

	m.instance.State = m.getState().String()

	ctx, cancel := context.WithTimeout(m.ctx, defaultTimeout)
	defer cancel()

	if err := m.opts.registry.Register(ctx, m.instance); err != nil {
		log.Fatalf("refresh cluster instance failed: %v", err)
	}
}

// 解注册服务实例
func (m *Mesh) deregisterServiceInstance() {
	ctx, cancel := context.WithTimeout(m.ctx, defaultTimeout)
	defer cancel()

	if err := m.opts.registry.Deregister(ctx, m.instance); err != nil {
		log.Errorf("deregister cluster instance failed: %v", err)
	}
}

// 获取状态
func (m *Mesh) getState() cluster.State {
	return cluster.State(m.state.Load())
}

// 执行钩子函数
func (m *Mesh) runHookFunc(hook cluster.Hook) {
	m.rw.RLock()

	if handlers, ok := m.hooks[hook]; ok {
		wg := &sync.WaitGroup{}
		wg.Add(len(handlers))

		for i := range handlers {
			handler := handlers[i]
			xcall.Go(func() {
				handler(m.proxy)
				wg.Done()
			})
		}

		m.rw.RUnlock()

		wg.Wait()
	} else {
		m.rw.RUnlock()
	}
}

// 添加钩子监听器
func (m *Mesh) addHookListener(hook cluster.Hook, handler HookHandler) {
	switch hook {
	case cluster.Destroy:
		m.rw.Lock()
		m.hooks[hook] = append(m.hooks[hook], handler)
		m.rw.Unlock()
	default:
		if m.getState() == cluster.Shut {
			m.hooks[hook] = append(m.hooks[hook], handler)
		} else {
			log.Warnf("server is working, can't add hook handler")
		}
	}
}

// 添加服务提供者
func (m *Mesh) addServiceProvider(name string, desc, provider any) {
	if m.getState() == cluster.Shut {
		m.services = append(m.services, &serviceEntity{
			name:     name,
			desc:     desc,
			provider: provider,
		})
	} else {
		log.Warnf("mesh server is working, can't add service provider")
	}
}

// 打印组件信息
func (m *Mesh) printInfo() {
	infos := make([]string, 0, 7)
	infos = append(infos, fmt.Sprintf("ID: %s", m.opts.id))
	infos = append(infos, fmt.Sprintf("Name: %s", m.Name()))
	infos = append(infos, fmt.Sprintf("Codec: %s", m.opts.codec.Name()))

	if m.opts.locator != nil {
		infos = append(infos, fmt.Sprintf("Locator: %s", m.opts.locator.Name()))
	} else {
		infos = append(infos, "Locator: -")
	}

	infos = append(infos, fmt.Sprintf("Registry: %s", m.opts.registry.Name()))

	if m.opts.encryptor != nil {
		infos = append(infos, fmt.Sprintf("Encryptor: %s", m.opts.encryptor.Name()))
	} else {
		infos = append(infos, "Encryptor: -")
	}

	infos = append(infos, fmt.Sprintf("Transporter: %s", m.opts.transporter.Name()))

	info.PrintBoxInfo("Mesh", infos...)
}
