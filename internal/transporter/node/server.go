package node

import (
	"context"
	"fmt"
	"github.com/develop-top/due/v2/cluster"
	"github.com/develop-top/due/v2/core/buffer"
	"github.com/develop-top/due/v2/errors"
	"github.com/develop-top/due/v2/internal/transporter/internal/codes"
	"github.com/develop-top/due/v2/internal/transporter/internal/protocol"
	"github.com/develop-top/due/v2/internal/transporter/internal/route"
	"github.com/develop-top/due/v2/internal/transporter/internal/server"
	"github.com/develop-top/due/v2/tracer"
	"github.com/develop-top/due/v2/utils/xtrace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Server struct {
	*server.Server
	provider Provider
}

func NewServer(addr string, provider Provider) (*Server, error) {
	serv, err := server.NewServer(&server.Options{Addr: addr})
	if err != nil {
		return nil, err
	}

	s := &Server{Server: serv, provider: provider}
	s.init()

	return s, nil
}

func (s *Server) init() {
	s.RegisterHandler(route.Trigger, s.trigger)
	s.RegisterHandler(route.Deliver, s.deliver)
	s.RegisterHandler(route.GetState, s.getState)
	s.RegisterHandler(route.SetState, s.setState)
}

// 携带链路追踪信息
func (s *Server) traceBuffer(ctx context.Context, route uint8, seq uint64, buf buffer.Buffer) buffer.Buffer {
	if !tracer.IsOpen {
		return protocol.EncodeBuffer(protocol.DataBit, route, seq, nil, buf)
	}

	traceCtx := protocol.MarshalSpanContext(trace.SpanContextFromContext(ctx))
	return protocol.EncodeBuffer(protocol.DataBit, route, seq, traceCtx, buf)
}

func (s *Server) startSpan(ctx context.Context, routeID uint8, attr ...attribute.KeyValue) (context.Context, trace.Span, func()) {
	if !tracer.IsOpen {
		return ctx, nil, func() {}
	}

	name := route.Name[routeID]

	ctx, span := xtrace.StartRPCServerSpan(ctx, fmt.Sprintf("node.server.%v", name),
		append([]attribute.KeyValue{
			tracer.RPCMessageTypeReceived,
		}, attr...)...,
	)

	return ctx, span, func() { span.End() }
}

// 触发事件
func (s *Server) trigger(ctx context.Context, conn *server.Conn, data []byte) error {
	ctx, span, end := s.startSpan(ctx, route.Trigger)
	defer end()

	seq, event, cid, uid, err := protocol.DecodeTriggerReq(data)
	if err != nil {
		return err
	}

	if span != nil {
		span.SetName(fmt.Sprintf("node.server.%v.%v", route.Name[route.Trigger], event.String()))
	}

	if conn.InsKind != cluster.Gate {
		return errors.ErrIllegalRequest
	}

	if err = s.provider.Trigger(ctx, conn.InsID, cid, uid, event); seq == 0 {
		if errors.Is(err, errors.ErrNotFoundSession) {
			return nil
		} else {
			return err
		}
	} else {
		return conn.Send(ctx, s.traceBuffer(ctx, route.Trigger, seq, protocol.EncodeTriggerRes(codes.ErrorToCode(err))))
	}
}

// 投递消息
func (s *Server) deliver(ctx context.Context, conn *server.Conn, data []byte) error {
	ctx, _, end := s.startSpan(ctx, route.Deliver)
	defer end()

	seq, cid, uid, message, err := protocol.DecodeDeliverReq(data)
	if err != nil {
		return err
	}

	var (
		gid string
		nid string
	)

	switch conn.InsKind {
	case cluster.Gate:
		gid = conn.InsID
	case cluster.Node:
		nid = conn.InsID
	default:
		return errors.ErrIllegalRequest
	}

	if err = s.provider.Deliver(ctx, gid, nid, cid, uid, message); seq == 0 {
		return err
	} else {
		return conn.Send(ctx, s.traceBuffer(ctx, route.Deliver, seq, protocol.EncodeDeliverRes(codes.ErrorToCode(err))))
	}
}

// 获取状态
func (s *Server) getState(ctx context.Context, conn *server.Conn, data []byte) error {
	ctx, _, end := s.startSpan(ctx, route.GetState)
	defer end()

	seq, err := protocol.DecodeSeq(buffer.NewReader(data))
	if err != nil {
		return err
	}

	state, err := s.provider.GetState(ctx)

	return conn.Send(ctx, s.traceBuffer(ctx, route.GetState, seq, protocol.EncodeGetStateRes(codes.ErrorToCode(err), state)))
}

// 设置状态
func (s *Server) setState(ctx context.Context, conn *server.Conn, data []byte) error {
	ctx, span, end := s.startSpan(ctx, route.SetState)
	defer end()

	seq, state, err := protocol.DecodeSetStateReq(data)
	if err != nil {
		return err
	}

	if span != nil {
		span.SetName(fmt.Sprintf("node.server.%v.%v", route.Name[route.SetState], state.String()))
	}

	err = s.provider.SetState(ctx, state)

	return conn.Send(ctx, s.traceBuffer(ctx, route.SetState, seq, protocol.EncodeSetStateRes(codes.ErrorToCode(err))))
}
