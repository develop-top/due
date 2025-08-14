/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2022/7/7 1:31 上午
 * @Desc: TODO
 */

package gate

import (
	"context"
	"time"

	"github.com/develop-top/due/v2/cluster"
	"github.com/develop-top/due/v2/etc"
	"github.com/develop-top/due/v2/locate"
	"github.com/develop-top/due/v2/log"
	"github.com/develop-top/due/v2/network"
	"github.com/develop-top/due/v2/registry"
	"github.com/develop-top/due/v2/utils/xuuid"
)

const (
	defaultName     = "gate"          // 默认名称
	defaultAddr     = ":0"            // 连接器监听地址
	defaultTimeout  = 3 * time.Second // 默认超时时间
	defaultDispatch = cluster.Random  // 默认的无状态路由分发策略
)

const (
	defaultIDKey       = "etc.cluster.gate.id"
	defaultNameKey     = "etc.cluster.gate.name"
	defaultAddrKey     = "etc.cluster.gate.addr"
	defaultTimeoutKey  = "etc.cluster.gate.timeout"
	defaultDispatchKey = "etc.cluster.gate.dispatch"
	defaultMetadataKey = "etc.cluster.gate.metadata"
)

type Option func(o *options)

type options struct {
	ctx      context.Context   // 上下文
	id       string            // 实例ID
	name     string            // 实例名称
	addr     string            // 监听地址
	timeout  time.Duration     // RPC调用超时时间
	server   network.Server    // 网关服务器
	locator  locate.Locator    // 用户定位器
	registry registry.Registry // 服务注册器
	dispatch cluster.Dispatch  // 无状态路由消息分发策略
	metadata map[string]any    // 元数据
}

func defaultOptions() *options {
	opts := &options{
		ctx:      context.Background(),
		name:     defaultName,
		addr:     defaultAddr,
		timeout:  defaultTimeout,
		dispatch: defaultDispatch,
		metadata: make(map[string]any),
	}

	if id := etc.Get(defaultIDKey).String(); id != "" {
		opts.id = id
	} else {
		opts.id = xuuid.UUID()
	}

	if name := etc.Get(defaultNameKey).String(); name != "" {
		opts.name = name
	}

	if addr := etc.Get(defaultAddrKey).String(); addr != "" {
		opts.addr = addr
	}

	if timeout := etc.Get(defaultTimeoutKey).Duration(); timeout > 0 {
		opts.timeout = timeout
	}

	if strategy := etc.Get(defaultDispatchKey).String(); strategy != "" {
		opts.dispatch = cluster.Dispatch(strategy)
	}

	if err := etc.Get(defaultMetadataKey).Scan(&opts.metadata); err != nil {
		log.Warnf("scan gate metadata failed: %v", err)
	}

	return opts
}

// WithID 设置实例ID
func WithID(id string) Option {
	return func(o *options) { o.id = id }
}

// WithName 设置实例名称
func WithName(name string) Option {
	return func(o *options) { o.name = name }
}

// WithContext 设置上下文
func WithContext(ctx context.Context) Option {
	return func(o *options) { o.ctx = ctx }
}

// WithServer 设置服务器
func WithServer(server network.Server) Option {
	return func(o *options) { o.server = server }
}

// WithTimeout 设置RPC调用超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(o *options) { o.timeout = timeout }
}

// WithLocator 设置用户定位器
func WithLocator(locator locate.Locator) Option {
	return func(o *options) { o.locator = locator }
}

// WithRegistry 设置服务注册器
func WithRegistry(r registry.Registry) Option {
	return func(o *options) { o.registry = r }
}

// WithDispatch 设置无状态路由消息分发策略
func WithDispatch(dispatch cluster.Dispatch) Option {
	return func(o *options) { o.dispatch = dispatch }
}

// WithMetadata 设置元数据
func WithMetadata(metadata map[string]any) Option {
	return func(o *options) { o.metadata = metadata }
}
