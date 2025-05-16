package gate

import (
	"context"
	"github.com/develop-top/due/v2/internal/transporter/internal/codes"
	"github.com/develop-top/due/v2/internal/transporter/internal/protocol"
	"github.com/develop-top/due/v2/internal/transporter/internal/route"
	"github.com/develop-top/due/v2/internal/transporter/internal/server"
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
	s.RegisterHandler(route.Bind, s.bind)
	s.RegisterHandler(route.Unbind, s.unbind)
	s.RegisterHandler(route.GetIP, s.getIP)
	s.RegisterHandler(route.Stat, s.stat)
	s.RegisterHandler(route.IsOnline, s.isOnline)
	s.RegisterHandler(route.Disconnect, s.disconnect)
	s.RegisterHandler(route.Push, s.push)
	s.RegisterHandler(route.Multicast, s.multicast)
	s.RegisterHandler(route.Broadcast, s.broadcast)
}

// 绑定用户
func (s *Server) bind(ctx context.Context, conn *server.Conn, data []byte) error {
	seq, cid, uid, err := protocol.DecodeBindReq(data)
	if err != nil {
		return err
	}

	if err = s.provider.Bind(ctx, cid, uid); seq == 0 {
		return err
	} else {
		return conn.Send(protocol.EncodeBindRes(seq, codes.ErrorToCode(err)))
	}
}

// 解绑用户
func (s *Server) unbind(ctx context.Context, conn *server.Conn, data []byte) error {
	seq, uid, err := protocol.DecodeUnbindReq(data)
	if err != nil {
		return err
	}

	if err = s.provider.Unbind(ctx, uid); seq == 0 {
		return err
	} else {
		return conn.Send(protocol.EncodeUnbindRes(seq, codes.ErrorToCode(err)))
	}
}

// 获取IP地址
func (s *Server) getIP(ctx context.Context, conn *server.Conn, data []byte) error {
	seq, kind, target, err := protocol.DecodeGetIPReq(data)
	if err != nil {
		return err
	}

	if ip, err := s.provider.GetIP(ctx, kind, target); seq == 0 {
		return err
	} else {
		return conn.Send(protocol.EncodeGetIPRes(seq, codes.ErrorToCode(err), ip))
	}
}

// 统计在线人数
func (s *Server) stat(ctx context.Context, conn *server.Conn, data []byte) error {
	seq, kind, err := protocol.DecodeStatReq(data)
	if err != nil {
		return err
	}

	if total, err := s.provider.Stat(ctx, kind); seq == 0 {
		return err
	} else {
		return conn.Send(protocol.EncodeStatRes(seq, codes.ErrorToCode(err), uint64(total)))
	}
}

// 检测用户是否在线
func (s *Server) isOnline(ctx context.Context, conn *server.Conn, data []byte) error {
	seq, kind, target, err := protocol.DecodeIsOnlineReq(data)
	if err != nil {
		return err
	}

	if isOnline, err := s.provider.IsOnline(ctx, kind, target); seq == 0 {
		return err
	} else {
		return conn.Send(protocol.EncodeIsOnlineRes(seq, codes.ErrorToCode(err), isOnline))
	}
}

// 断开连接
func (s *Server) disconnect(ctx context.Context, conn *server.Conn, data []byte) error {
	seq, kind, target, force, err := protocol.DecodeDisconnectReq(data)
	if err != nil {
		return err
	}

	if err = s.provider.Disconnect(ctx, kind, target, force); seq == 0 {
		return err
	} else {
		return conn.Send(protocol.EncodeDisconnectRes(seq, codes.ErrorToCode(err)))
	}
}

// 推送单个消息
func (s *Server) push(ctx context.Context, conn *server.Conn, data []byte) error {
	seq, kind, target, message, err := protocol.DecodePushReq(data)
	if err != nil {
		return err
	}

	if err = s.provider.Push(ctx, kind, target, message); seq == 0 {
		return err
	} else {
		return conn.Send(protocol.EncodePushRes(seq, codes.ErrorToCode(err)))
	}
}

// 推送组播消息
func (s *Server) multicast(ctx context.Context, conn *server.Conn, data []byte) error {
	seq, kind, targets, message, err := protocol.DecodeMulticastReq(data)
	if err != nil {
		return err
	}

	if total, err := s.provider.Multicast(ctx, kind, targets, message); seq == 0 {
		return err
	} else {
		return conn.Send(protocol.EncodeMulticastRes(seq, codes.ErrorToCode(err), uint64(total)))
	}
}

// 推送广播消息
func (s *Server) broadcast(ctx context.Context, conn *server.Conn, data []byte) error {
	seq, kind, message, err := protocol.DecodeBroadcastReq(data)
	if err != nil {
		return err
	}

	if total, err := s.provider.Broadcast(ctx, kind, message); seq == 0 {
		return err
	} else {
		return conn.Send(protocol.EncodeBroadcastRes(seq, codes.ErrorToCode(err), uint64(total)))
	}
}

// 获取状态
func (s *Server) getState(conn *server.Conn, data []byte) error {
	seq, err := protocol.DecodeGetStateReq(data)
	if err != nil {
		return err
	}

	state, err := s.provider.GetState()

	return conn.Send(protocol.EncodeGetStateRes(seq, codes.ErrorToCode(err), state))
}

// 设置状态
func (s *Server) setState(conn *server.Conn, data []byte) error {
	seq, state, err := protocol.DecodeSetStateReq(data)
	if err != nil {
		return err
	}

	err = s.provider.SetState(state)

	return conn.Send(protocol.EncodeSetStateRes(seq, codes.ErrorToCode(err)))
}
