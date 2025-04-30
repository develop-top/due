package xtrace

import (
	"context"

	xtrace "github.com/develop-top/due/v2/tracer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	gcodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryTracingGrpcServerInterceptor is a grpc.UnaryServerInterceptor for opentelemetry.
func UnaryTracingGrpcServerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (any, error) {
	ctx, span := startServerSpan(ctx, info.FullMethod)
	defer span.End()

	xtrace.MessageReceived.Event(ctx, 1, req)
	resp, err := handler(ctx, req)
	if err != nil {
		s, ok := status.FromError(err)
		if ok {
			span.SetStatus(codes.Error, s.Message())
			span.SetAttributes(xtrace.StatusCodeAttr(s.Code()))
			xtrace.MessageSent.Event(ctx, 1, s.Proto())
		} else {
			span.SetStatus(codes.Error, err.Error())
		}
		return nil, err
	}

	span.SetAttributes(xtrace.StatusCodeAttr(gcodes.OK))
	xtrace.MessageSent.Event(ctx, 1, resp)

	return resp, nil
}

// StreamTracingGrpcServerInterceptor returns a grpc.StreamServerInterceptor for opentelemetry.
func StreamTracingGrpcServerInterceptor(svr any, ss grpc.ServerStream, info *grpc.StreamServerInfo,
	handler grpc.StreamHandler) error {
	ctx, span := startServerSpan(ss.Context(), info.FullMethod)
	defer span.End()

	if err := handler(svr, wrapServerStream(ctx, ss)); err != nil {
		s, ok := status.FromError(err)
		if ok {
			span.SetStatus(codes.Error, s.Message())
			span.SetAttributes(xtrace.StatusCodeAttr(s.Code()))
		} else {
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}

	span.SetAttributes(xtrace.StatusCodeAttr(gcodes.OK))
	return nil
}

// serverStream wraps around the embedded grpc.ServerStream,
// and intercepts the RecvMsg and SendMsg method call.
type serverStream struct {
	grpc.ServerStream
	ctx               context.Context
	receivedMessageID int
	sentMessageID     int
}

func (w *serverStream) Context() context.Context {
	return w.ctx
}

func (w *serverStream) RecvMsg(m any) error {
	err := w.ServerStream.RecvMsg(m)
	if err == nil {
		w.receivedMessageID++
		xtrace.MessageReceived.Event(w.Context(), w.receivedMessageID, m)
	}

	return err
}

func (w *serverStream) SendMsg(m any) error {
	err := w.ServerStream.SendMsg(m)
	w.sentMessageID++
	xtrace.MessageSent.Event(w.Context(), w.sentMessageID, m)

	return err
}

func startServerSpan(ctx context.Context, method string) (context.Context, trace.Span) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	bags, spanCtx := Extract(ctx, otel.GetTextMapPropagator(), &md)
	ctx = baggage.ContextWithBaggage(ctx, bags)
	tr := otel.Tracer(xtrace.TraceName)
	name, attr := xtrace.SpanInfo(method, xtrace.PeerFromCtx(ctx))

	return tr.Start(trace.ContextWithRemoteSpanContext(ctx, spanCtx), name,
		trace.WithSpanKind(trace.SpanKindServer), trace.WithAttributes(attr...))
}

// wrapServerStream wraps the given grpc.ServerStream with the given context.
func wrapServerStream(ctx context.Context, ss grpc.ServerStream) *serverStream {
	return &serverStream{
		ServerStream: ss,
		ctx:          ctx,
	}
}
