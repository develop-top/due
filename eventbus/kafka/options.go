package kafka

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/develop-top/due/v2/etc"
)

const (
	defaultAddr   = "127.0.0.1:9092"
	defaultPrefix = "due"
)

const (
	defaultAddrsKey   = "etc.eventbus.kafka.addrs"
	defaultPrefixKey  = "etc.eventbus.kafka.prefix"
	defaultVersionKey = "etc.eventbus.kafka.version"
)

type Option func(o *options)

type options struct {
	ctx context.Context

	// 客户端连接地址
	// 内建客户端配置，默认为[]string{"127.0.0.1:9092"}
	addrs []string

	// Kafka版本，默认为无版本
	version string

	// 前缀
	// key前缀，默认为due
	prefix string

	// 客户端
	// 外部客户端配置，存在外部客户端时，优先使用外部客户端，默认为nil
	client sarama.Client
}

func defaultOptions() *options {
	return &options{
		ctx:     context.Background(),
		addrs:   etc.Get(defaultAddrsKey, []string{defaultAddr}).Strings(),
		prefix:  etc.Get(defaultPrefixKey, defaultPrefix).String(),
		version: etc.Get(defaultVersionKey).String(),
	}
}

// WithContext 设置上下文
func WithContext(ctx context.Context) Option {
	return func(o *options) { o.ctx = ctx }
}

// WithAddrs 设置连接地址
func WithAddrs(addrs ...string) Option {
	return func(o *options) { o.addrs = addrs }
}

// WithPrefix 设置前缀
func WithPrefix(prefix string) Option {
	return func(o *options) { o.prefix = prefix }
}

// WithVersion 设置Kafka版本
func WithVersion(version string) Option {
	return func(o *options) { o.version = version }
}

// WithClient 设置外部客户端
func WithClient(client sarama.Client) Option {
	return func(o *options) { o.client = client }
}
