package gate

import (
	"context"
	"fmt"
	"sync/atomic"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/develop-top/due/v2/cluster"
	"github.com/develop-top/due/v2/core/buffer"
	"github.com/develop-top/due/v2/errors"
	"github.com/develop-top/due/v2/internal/transporter/internal/client"
	"github.com/develop-top/due/v2/internal/transporter/internal/codes"
	"github.com/develop-top/due/v2/internal/transporter/internal/protocol"
	"github.com/develop-top/due/v2/internal/transporter/internal/route"
	"github.com/develop-top/due/v2/session"
	"github.com/develop-top/due/v2/tracer"
)

type Client struct {
	seq uint64
	cli *client.Client
}

func NewClient(cli *client.Client) *Client {
	return &Client{
		cli: cli,
	}
}

// 携带链路追踪信息
func (c *Client) traceBuffer(ctx context.Context, routeID uint8, seq uint64, buf buffer.Buffer, attr ...attribute.KeyValue) (
	context.Context, func(), buffer.Buffer) {
	if !tracer.IsOpen {
		return ctx, func() {}, protocol.EncodeBuffer(protocol.DataBit, routeID, seq, nil, buf)
	}

	name := route.Name[routeID]

	traceCtx := protocol.MarshalSpanContext(trace.SpanContextFromContext(ctx))
	buf = protocol.EncodeBuffer(protocol.DataBit, routeID, seq, traceCtx, buf)

	myCtx, span := tracer.NewSpan(ctx, fmt.Sprintf("gate.client.%s", name),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			append([]attribute.KeyValue{
				tracer.RPCMessageTypeSent,
				tracer.RPCMessageIDKey.String(name),
				tracer.RPCMessageCompressedSizeKey.Int(buf.Len()),
				tracer.InstanceKind.String(c.cli.Opts.InsKind.String()),
				tracer.InstanceID.String(c.cli.Opts.InsID),
				tracer.ServerIP.String(c.cli.Opts.Addr),
			}, attr...)...))

	return myCtx, func() { span.End() }, buf
}

// Bind 绑定用户与连接
func (c *Client) Bind(ctx context.Context, cid, uid int64) (bool, error) {
	seq := c.doGenSequence()

	ctx, end, buf := c.traceBuffer(ctx, route.Bind, seq, protocol.EncodeBindReq(cid, uid))
	defer end()

	res, err := c.cli.Call(ctx, seq, buf)
	if err != nil {
		return false, err
	}

	code, err := protocol.DecodeBindRes(res)
	if err != nil {
		return false, err
	}

	return code == codes.NotFoundSession, nil
}

// Unbind 解绑用户与连接
func (c *Client) Unbind(ctx context.Context, uid int64) (bool, error) {
	seq := c.doGenSequence()

	ctx, end, buf := c.traceBuffer(ctx, route.Unbind, seq, protocol.EncodeUnbindReq(uid))
	defer end()

	res, err := c.cli.Call(ctx, seq, buf)
	if err != nil {
		return false, err
	}

	code, err := protocol.DecodeUnbindRes(res)
	if err != nil {
		return false, err
	}

	return code == codes.NotFoundSession, nil
}

// GetIP 获取客户端IP
func (c *Client) GetIP(ctx context.Context, kind session.Kind, target int64) (string, bool, error) {
	seq := c.doGenSequence()

	ctx, end, buf := c.traceBuffer(ctx, route.GetIP, seq, protocol.EncodeGetIPReq(kind, target))
	defer end()

	res, err := c.cli.Call(ctx, seq, buf)
	if err != nil {
		return "", false, err
	}

	code, ip, err := protocol.DecodeGetIPRes(res)
	if err != nil {
		return "", false, err
	}

	return ip, code == codes.NotFoundSession, nil
}

// Stat 推送广播消息
func (c *Client) Stat(ctx context.Context, kind session.Kind) (int64, error) {
	seq := c.doGenSequence()

	ctx, end, buf := c.traceBuffer(ctx, route.Stat, seq, protocol.EncodeStatReq(kind))
	defer end()

	res, err := c.cli.Call(ctx, seq, buf)
	if err != nil {
		return 0, err
	}

	_, total, err := protocol.DecodeStatRes(res)

	return int64(total), err
}

// IsOnline 检测是否在线
func (c *Client) IsOnline(ctx context.Context, kind session.Kind, target int64) (bool, bool, error) {
	seq := c.doGenSequence()

	ctx, end, buf := c.traceBuffer(ctx, route.IsOnline, seq, protocol.EncodeIsOnlineReq(kind, target))
	defer end()

	res, err := c.cli.Call(ctx, seq, buf)
	if err != nil {
		return false, false, err
	}

	code, isOnline, err := protocol.DecodeIsOnlineRes(res)
	if err != nil {
		return false, false, err
	}

	return code == codes.NotFoundSession, isOnline, nil
}

// Disconnect 断开连接
func (c *Client) Disconnect(ctx context.Context, kind session.Kind, target int64, force bool) error {
	ctx, end, buf := c.traceBuffer(ctx, route.Disconnect, 0, protocol.EncodeDisconnectReq(kind, target, force))
	defer end()

	if force {
		return c.cli.Send(ctx, buf)
	} else {
		return c.cli.Send(ctx, buf, target)
	}
}

// Push 推送消息（异步）
func (c *Client) Push(ctx context.Context, kind session.Kind, target int64, message buffer.Buffer) error {
	ctx, end, buf := c.traceBuffer(ctx, route.Push, 0, protocol.EncodePushReq(kind, target, message))
	defer end()
	return c.cli.Send(ctx, buf, target)
}

// Multicast 推送组播消息（异步）
func (c *Client) Multicast(ctx context.Context, kind session.Kind, targets []int64, message buffer.Buffer) error {
	ctx, end, buf := c.traceBuffer(ctx, route.Multicast, 0, protocol.EncodeMulticastReq(kind, targets, message))
	defer end()
	return c.cli.Send(ctx, buf)
}

// Broadcast 推送广播消息（异步）
func (c *Client) Broadcast(ctx context.Context, kind session.Kind, message buffer.Buffer) error {
	ctx, end, buf := c.traceBuffer(ctx, route.Broadcast, 0, protocol.EncodeBroadcastReq(kind, message))
	defer end()
	return c.cli.Send(ctx, buf)
}

// Publish 发布频道消息（异步）
func (c *Client) Publish(ctx context.Context, channel string, message buffer.Buffer) error {
	if len(channel) > 1<<8-1 {
		return errors.ErrInvalidArgument
	}

	ctx, end, buf := c.traceBuffer(ctx, route.Publish, 0, protocol.EncodePublishReq(channel, message))
	defer end()

	return c.cli.Send(ctx, buf)
}

// Subscribe 订阅频道
func (c *Client) Subscribe(ctx context.Context, kind session.Kind, targets []int64, channel string) error {
	if len(channel) > 1<<8-1 {
		return errors.ErrInvalidArgument
	}

	seq := c.doGenSequence()

	ctx, end, buf := c.traceBuffer(ctx, route.Subscribe, seq, protocol.EncodeSubscribeReq(kind, targets, channel))
	defer end()

	res, err := c.cli.Call(ctx, seq, buf)
	if err != nil {
		return err
	}

	code, err := protocol.DecodeSubscribeRes(res)
	if err != nil {
		return err
	}

	return codes.CodeToError(code)
}

// Unsubscribe 取消订阅频道
func (c *Client) Unsubscribe(ctx context.Context, kind session.Kind, targets []int64, channel string) error {
	if len(channel) > 1<<8-1 {
		return errors.ErrInvalidArgument
	}

	seq := c.doGenSequence()

	ctx, end, buf := c.traceBuffer(ctx, route.Unsubscribe, seq, protocol.EncodeUnsubscribeReq(kind, targets, channel))
	defer end()

	res, err := c.cli.Call(ctx, seq, buf)
	if err != nil {
		return err
	}

	code, err := protocol.DecodeUnsubscribeRes(res)
	if err != nil {
		return err
	}

	return codes.CodeToError(code)
}

// GetState 获取状态
func (c *Client) GetState(ctx context.Context) (cluster.State, error) {
	seq := c.doGenSequence()

	ctx, end, buf := c.traceBuffer(ctx, route.GetState, seq, nil)
	defer end()

	res, err := c.cli.Call(ctx, seq, buf)
	if err != nil {
		return 0, err
	}

	code, state, err := protocol.DecodeGetStateRes(res)
	if err != nil {
		return 0, err
	}

	return state, codes.CodeToError(code)
}

// SetState 设置状态
func (c *Client) SetState(ctx context.Context, state cluster.State) error {
	seq := c.doGenSequence()

	ctx, end, buf := c.traceBuffer(ctx, route.SetState, seq, protocol.EncodeSetStateReq(state))
	defer end()

	res, err := c.cli.Call(ctx, seq, buf)
	if err != nil {
		return err
	}

	code, err := protocol.DecodeSetStateRes(res)
	if err != nil {
		return err
	}

	return codes.CodeToError(code)
}

// 生成序列号，规避生成序列号为0的编号
func (c *Client) doGenSequence() (seq uint64) {
	for {
		if seq = atomic.AddUint64(&c.seq, 1); seq != 0 {
			return
		}
	}
}
