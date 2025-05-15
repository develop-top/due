package xtrace

import (
	"context"
	xtrace "github.com/develop-top/due/v2/tracer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

func StartRPCClientSpan(ctx context.Context, name string, attr ...attribute.KeyValue) (context.Context, trace.Span) {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	tr := otel.Tracer(xtrace.TraceName)
	ctx, span := tr.Start(ctx, name,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(xtrace.RPCSystemTCP),
		trace.WithAttributes(attr...))
	Inject(ctx, otel.GetTextMapPropagator(), &md)
	ctx = metadata.NewOutgoingContext(ctx, md)

	return ctx, span
}

func StartRPCServerSpan(ctx context.Context, name string, attr ...attribute.KeyValue) (context.Context, trace.Span) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	bags, spanCtx := Extract(ctx, otel.GetTextMapPropagator(), &md)
	ctx = baggage.ContextWithBaggage(ctx, bags)
	tr := otel.Tracer(xtrace.TraceName)

	return tr.Start(trace.ContextWithRemoteSpanContext(ctx, spanCtx), name,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(xtrace.RPCSystemTCP),
		trace.WithAttributes(attr...))
}
