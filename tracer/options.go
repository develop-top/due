package tracer

import "github.com/develop-top/due/v2/etc"

const (
	defaultTraceName = "due"
	defaultName      = "due"
	defaultEndpoint  = "http://127.0.0.1:14268/api/traces"
	defaultSampler   = 1.0
	defaultBatcher   = "jaeger"
)

const (
	defaultTraceNameKey = "etc.opentracing.traceName"
	defaultNameKey      = "etc.opentracing.name"
	defaultEndpointKey  = "etc.opentracing.endpoint"
	defaultSamplerKey   = "etc.opentracing.sampler"
	defaultBatcherKey   = "etc.opentracing.batcher"
	defaultDisableKey   = "etc.opentracing.disabled"
	defaultDueReportKey = "etc.opentracing.dueReport"
)

var TraceName = "due" // 系统名称

type Option func(*Options)

type Options struct {
	TraceName string  // 系统名称
	Name      string  // 服务名称
	Endpoint  string  // 服务地址
	Sampler   float64 // 采集频率
	Batcher   string  // jaeger|zipkin|otlpgrpc|otlphttp|file
	// OtlpHeaders represents the headers for OTLP gRPC or HTTP transport.
	// For example:
	//  uptrace-dsn: 'http://project2_secret_token@localhost:14317/2'
	OtlpHeaders map[string]string `json:",optional"`
	// OtlpHttpPath represents the path for OTLP HTTP transport.
	// For example
	// /v1/traces
	OtlpHttpPath string `json:",optional"`
	// OtlpHttpSecure represents the scheme to use for OTLP HTTP transport.
	OtlpHttpSecure bool `json:",optional"`
	// Disabled indicates whether StartAgent starts the agent.
	Disabled  bool `json:",optional"`
	DueReport bool `json:",optional"`
}

func defaultOptions() *Options {
	return &Options{
		TraceName: etc.Get(defaultTraceNameKey, defaultTraceName).String(),
		Name:      etc.Get(defaultNameKey, defaultName).String(),
		Endpoint:  etc.Get(defaultEndpointKey, defaultEndpoint).String(),
		Sampler:   etc.Get(defaultSamplerKey, defaultSampler).Float64(),
		Batcher:   etc.Get(defaultBatcherKey, defaultBatcher).String(),
		Disabled:  etc.Get(defaultDisableKey, true).Bool(),   // 默认禁用
		DueReport: etc.Get(defaultDueReportKey, true).Bool(), // 上报due链路追踪数据
	}
}

func WithTraceName(name string) Option {
	return func(o *Options) {
		o.TraceName = name
	}
}

func WithName(name string) Option {
	return func(o *Options) {
		o.Name = name
	}
}

func WithEndpoint(endpoint string) Option {
	return func(o *Options) {
		o.Endpoint = endpoint
	}
}

func WithSampler(sampler float64) Option {
	return func(o *Options) {
		o.Sampler = sampler
	}
}

func WithBatcher(batcher string) Option {
	return func(o *Options) {
		o.Batcher = batcher
	}
}

func WithOtlpHeaders(headers map[string]string) Option {
	return func(o *Options) {
		o.OtlpHeaders = headers
	}
}

func WithOtlpHttpPath(path string) Option {
	return func(o *Options) {
		o.OtlpHttpPath = path
	}
}

func WithOtlpHttpSecure(secure bool) Option {
	return func(o *Options) {
		o.OtlpHttpSecure = secure
	}
}

func WithDisabled(disabled bool) Option {
	return func(o *Options) {
		o.Disabled = disabled
	}
}
