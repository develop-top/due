package gate

import (
	"context"
	"fmt"

	"github.com/develop-top/due/v2/core/buffer"
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

type ServerOptions = server.Options

func NewServer(provider Provider, opts *ServerOptions) (*Server, error) {
	serv, err := server.NewServer(opts)
	if err != nil {
		return nil, err
	}

	s := &Server{Server: serv, provider: provider}
	s.init()

	return s, nil
}

func (s *Server) init() {
	s.RegisterHandler(route.Bind, s.bind)
	s.RegisterHandler(route.Unbind, s.unbind)
	s.RegisterHandler(route.GetIP, s.getIP)
	s.RegisterHandler(route.Stat, s.stat)
	s.RegisterHandler(route.IsOnline, s.isOnline)
	s.RegisterHandler(route.Disconnect, s.disconnect)
	s.RegisterHandler(route.Push, s.push)
	s.RegisterHandler(route.Multicast, s.multicast)
	s.RegisterHandler(route.Broadcast, s.broadcast)
	s.RegisterHandler(route.Publish, s.publish)
	s.RegisterHandler(route.Subscribe, s.subscribe)
	s.RegisterHandler(route.Unsubscribe, s.unsubscribe)
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

func (s *Server) startSpan(ctx context.Context, r uint8, attr ...attribute.KeyValue) (context.Context, trace.Span, func()) {
	if !tracer.IsOpen {
		return ctx, nil, func() {}
	}

	ctx, span := xtrace.StartRPCServerSpan(ctx, fmt.Sprintf("gate.server.%s", route.Name[r]),
		append([]attribute.KeyValue{
			tracer.RPCMessageTypeReceived,
		}, attr...)...,
	)

	return ctx, span, func() { span.End() }
}

// 绑定用户
func (s *Server) bind(ctx context.Context, conn *server.Conn, data []byte) error {
	ctx, _, end := s.startSpan(ctx, route.Bind)
	defer end()

	seq, cid, uid, err := protocol.DecodeBindReq(data)
	if err != nil {
		return err
	}

	if err = s.provider.Bind(ctx, cid, uid); seq == 0 {
		return err
	} else {
		return conn.Send(ctx, s.traceBuffer(ctx, route.Bind, seq, protocol.EncodeBindRes(codes.ErrorToCode(err))))
	}
}

// 解绑用户
func (s *Server) unbind(ctx context.Context, conn *server.Conn, data []byte) error {
	ctx, _, end := s.startSpan(ctx, route.Unbind)
	defer end()

	seq, uid, err := protocol.DecodeUnbindReq(data)
	if err != nil {
		return err
	}

	if err = s.provider.Unbind(ctx, uid); seq == 0 {
		return err
	} else {
		return conn.Send(ctx, s.traceBuffer(ctx, route.Unbind, seq, protocol.EncodeUnbindRes(codes.ErrorToCode(err))))
	}
}

// 获取IP地址
func (s *Server) getIP(ctx context.Context, conn *server.Conn, data []byte) error {
	ctx, _, end := s.startSpan(ctx, route.GetIP)
	defer end()

	seq, kind, target, err := protocol.DecodeGetIPReq(data)
	if err != nil {
		return err
	}

	if ip, err := s.provider.GetIP(ctx, kind, target); seq == 0 {
		return err
	} else {
		return conn.Send(ctx, s.traceBuffer(ctx, route.GetIP, seq, protocol.EncodeGetIPRes(codes.ErrorToCode(err), ip)))
	}
}

// 统计在线人数
func (s *Server) stat(ctx context.Context, conn *server.Conn, data []byte) error {
	ctx, _, end := s.startSpan(ctx, route.Stat)
	defer end()

	seq, kind, err := protocol.DecodeStatReq(data)
	if err != nil {
		return err
	}

	if total, err := s.provider.Stat(ctx, kind); seq == 0 {
		return err
	} else {
		return conn.Send(ctx, s.traceBuffer(ctx, route.Stat, seq, protocol.EncodeStatRes(codes.ErrorToCode(err), uint64(total))))
	}
}

// 检测用户是否在线
func (s *Server) isOnline(ctx context.Context, conn *server.Conn, data []byte) error {
	ctx, _, end := s.startSpan(ctx, route.IsOnline)
	defer end()

	seq, kind, target, err := protocol.DecodeIsOnlineReq(data)
	if err != nil {
		return err
	}

	if isOnline, err := s.provider.IsOnline(ctx, kind, target); seq == 0 {
		return err
	} else {
		return conn.Send(ctx, s.traceBuffer(ctx, route.IsOnline, seq, protocol.EncodeIsOnlineRes(codes.ErrorToCode(err), isOnline)))
	}
}

// 断开连接
func (s *Server) disconnect(ctx context.Context, conn *server.Conn, data []byte) error {
	ctx, _, end := s.startSpan(ctx, route.Disconnect)
	defer end()

	seq, kind, target, force, err := protocol.DecodeDisconnectReq(data)
	if err != nil {
		return err
	}

	if err = s.provider.Disconnect(ctx, kind, target, force); seq == 0 {
		return err
	} else {
		return conn.Send(ctx, s.traceBuffer(ctx, route.Disconnect, seq, protocol.EncodeDisconnectRes(codes.ErrorToCode(err))))
	}
}

// 推送单个消息
func (s *Server) push(ctx context.Context, conn *server.Conn, data []byte) error {
	ctx, _, end := s.startSpan(ctx, route.Push)
	defer end()

	seq, kind, target, message, err := protocol.DecodePushReq(data)
	if err != nil {
		return err
	}

	if err = s.provider.Push(ctx, kind, target, message); seq == 0 {
		return err
	} else {
		return conn.Send(ctx, s.traceBuffer(ctx, route.Push, seq, protocol.EncodePushRes(codes.ErrorToCode(err))))
	}
}

// 推送组播消息
func (s *Server) multicast(ctx context.Context, conn *server.Conn, data []byte) error {
	ctx, _, end := s.startSpan(ctx, route.Multicast)
	defer end()

	seq, kind, targets, message, err := protocol.DecodeMulticastReq(data)
	if err != nil {
		return err
	}

	if total, err := s.provider.Multicast(ctx, kind, targets, message); seq == 0 {
		return err
	} else {
		return conn.Send(ctx, s.traceBuffer(ctx, route.Multicast, seq, protocol.EncodeMulticastRes(codes.ErrorToCode(err), uint64(total))))
	}
}

// 推送广播消息
func (s *Server) broadcast(ctx context.Context, conn *server.Conn, data []byte) error {
	ctx, _, end := s.startSpan(ctx, route.Broadcast)
	defer end()

	seq, kind, message, err := protocol.DecodeBroadcastReq(data)
	if err != nil {
		return err
	}

	if total, err := s.provider.Broadcast(ctx, kind, message); seq == 0 {
		return err
	} else {
		return conn.Send(ctx, s.traceBuffer(ctx, route.Broadcast, seq, protocol.EncodeBroadcastRes(codes.ErrorToCode(err), uint64(total))))
	}
}

// 发布频道消息
func (s *Server) publish(ctx context.Context, conn *server.Conn, data []byte) error {
	ctx, _, end := s.startSpan(ctx, route.Publish)
	defer end()

	seq, channel, message, err := protocol.DecodePublishReq(data)
	if err != nil {
		return err
	}

	if total := s.provider.Publish(ctx, channel, message); seq == 0 {
		return nil
	} else {
		return conn.Send(ctx, s.traceBuffer(ctx, route.Publish, seq, protocol.EncodePublishRes(uint64(total))))
	}
}

// 订阅频道
func (s *Server) subscribe(ctx context.Context, conn *server.Conn, data []byte) error {
	ctx, _, end := s.startSpan(ctx, route.Subscribe)
	defer end()

	seq, kind, targets, channel, err := protocol.DecodeSubscribeReq(data)
	if err != nil {
		return err
	}

	if err = s.provider.Subscribe(ctx, kind, targets, channel); seq == 0 {
		return err
	} else {
		return conn.Send(ctx, s.traceBuffer(ctx, route.Subscribe, seq, protocol.EncodeSubscribeRes(codes.ErrorToCode(err))))
	}
}

// 取消订阅频道
func (s *Server) unsubscribe(ctx context.Context, conn *server.Conn, data []byte) error {
	ctx, _, end := s.startSpan(ctx, route.Unsubscribe)
	defer end()

	seq, kind, targets, channel, err := protocol.DecodeUnsubscribeReq(data)
	if err != nil {
		return err
	}

	if err = s.provider.Unsubscribe(ctx, kind, targets, channel); seq == 0 {
		return err
	} else {
		return conn.Send(ctx, s.traceBuffer(ctx, route.Unsubscribe, seq, protocol.EncodeUnsubscribeRes(codes.ErrorToCode(err))))
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
	ctx, _, end := s.startSpan(ctx, route.SetState)
	defer end()

	seq, state, err := protocol.DecodeSetStateReq(data)
	if err != nil {
		return err
	}

	if err = s.provider.SetState(ctx, state); seq == 0 {
		return err
	} else {
		return conn.Send(ctx, s.traceBuffer(ctx, route.SetState, seq, protocol.EncodeSetStateRes(codes.ErrorToCode(err))))
	}

}
