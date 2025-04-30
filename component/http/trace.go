package http

import (
	"github.com/develop-top/due/v2/mode"
	"github.com/develop-top/due/v2/tracer"
	"github.com/develop-top/due/v2/utils/xconv"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"slices"
)

type (
	// TraceOption defines the method to customize an traceOptions.
	TraceOption func(options *traceOptions)

	// traceOptions is TraceHandler options.
	traceOptions struct {
		traceIgnorePaths []string
	}
)

// TraceHttpHandler return a middleware that process the opentelemetry.
func TraceHttpHandler(serviceName, path string, opts ...TraceOption) func(ctx Context) error {
	var options traceOptions
	for _, opt := range opts {
		opt(&options)
	}

	ignorePaths := []string{}
	ignorePaths = append(ignorePaths, options.traceIgnorePaths...)

	return func(ctx Context) error {
		tr := otel.Tracer(tracer.TraceName)
		propagator := otel.GetTextMapPropagator()

		r := ctx.Request()
		spanName := path
		if len(spanName) == 0 {
			spanName = string(r.URI().Path())
		}

		if slices.Index(ignorePaths, spanName) != -1 {
			return ctx.Next()
		}

		ctxReq := propagator.Extract(ctx.Context(), &RequestHeaderCarrier{header: &r.Header})
		spanCtx, span := tr.Start(
			ctxReq,
			spanName,
			oteltrace.WithSpanKind(oteltrace.SpanKindServer),
			oteltrace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(
				serviceName, spanName, ctx.StdRequest())...),
		)
		if mode.IsDebugMode() {
			span.SetAttributes(tracer.HttpRequestHeaderKey.String(xconv.String(ctx.GetReqHeaders())), tracer.HttpRequestBodyKey.String(string(r.Body())))
		}
		defer func() {
			if mode.IsDebugMode() {
				span.SetAttributes(tracer.HttpResponseKey.String(ctx.Response().String()))
			}
			span.SetAttributes(semconv.HTTPAttributesFromHTTPStatusCode(ctx.Response().StatusCode())...)
			span.SetStatus(semconv.SpanStatusFromHTTPStatusCodeAndSpanKind(
				ctx.Response().StatusCode(), oteltrace.SpanKindServer))
			span.End()
		}()

		// convenient for tracking error messages
		propagator.Inject(spanCtx, &ResponseHeaderCarrier{header: &ctx.Response().Header})

		// propagate spanCtx down
		ctx.SetContext(spanCtx)
		return ctx.Next()
	}
}

// WithTraceIgnorePaths specifies the traceIgnorePaths option for TraceHandler.
func WithTraceIgnorePaths(traceIgnorePaths []string) TraceOption {
	return func(options *traceOptions) {
		options.traceIgnorePaths = append(options.traceIgnorePaths, traceIgnorePaths...)
	}
}

type RequestHeaderCarrier struct {
	header *fasthttp.RequestHeader
}

// Get returns the value associated with the passed key.
func (c *RequestHeaderCarrier) Get(key string) string {
	val := c.header.Peek(key)
	if len(val) == 0 {
		return ""
	}
	return string(val)
}

// Set stores the key-value pair.
func (c *RequestHeaderCarrier) Set(key string, value string) {
	c.header.Set(key, value)
}

// Keys lists the keys stored in this carrier.
func (c *RequestHeaderCarrier) Keys() []string {
	keys := make([]string, 0)
	c.header.VisitAll(func(key, value []byte) {
		keys = append(keys, string(key))
	})
	return keys
}

type ResponseHeaderCarrier struct {
	header *fasthttp.ResponseHeader
}

// Get returns the value associated with the passed key.
func (c *ResponseHeaderCarrier) Get(key string) string {
	val := c.header.Peek(key)
	if len(val) == 0 {
		return ""
	}
	return string(val)
}

// Set stores the key-value pair.
func (c *ResponseHeaderCarrier) Set(key string, value string) {
	c.header.Set(key, value)
}

// Keys lists the keys stored in this carrier.
func (c *ResponseHeaderCarrier) Keys() []string {
	keys := make([]string, 0)
	c.header.VisitAll(func(key, value []byte) {
		keys = append(keys, string(key))
	})
	return keys
}
