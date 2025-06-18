package xtrace

import (
	"context"
	xtrace "github.com/develop-top/due/v2/tracer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func StartRPCServerSpan(ctx context.Context, name string, attr ...attribute.KeyValue) (context.Context, trace.Span) {
	tr := otel.Tracer(xtrace.TraceName)
	return tr.Start(ctx, name,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(xtrace.RPCSystemTCP),
		trace.WithAttributes(attr...))
}
