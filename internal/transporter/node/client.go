package node

import (
	"context"
	"fmt"
	"sync/atomic"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/develop-top/due/v2/cluster"
	"github.com/develop-top/due/v2/core/buffer"
	"github.com/develop-top/due/v2/internal/transporter/internal/client"
	"github.com/develop-top/due/v2/internal/transporter/internal/codes"
	"github.com/develop-top/due/v2/internal/transporter/internal/protocol"
	"github.com/develop-top/due/v2/internal/transporter/internal/route"
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

func (c *Client) traceBuffer(ctx context.Context, routeID uint8, seq uint64, buf buffer.Buffer, attr ...attribute.KeyValue) (
	context.Context, trace.Span, func(), buffer.Buffer) {
	if !tracer.IsOpen {
		return ctx, nil, func() {}, protocol.EncodeBuffer(protocol.DataBit, routeID, seq, nil, buf)
	}

	routeName := route.Name[routeID]

	traceCtx := protocol.MarshalSpanContext(trace.SpanContextFromContext(ctx))
	buf = protocol.EncodeBuffer(protocol.DataBit, routeID, seq, traceCtx, buf)

	myCtx, span := tracer.NewSpan(ctx, fmt.Sprintf("node.client.%s", routeName),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			append([]attribute.KeyValue{
				tracer.RPCMessageTypeSent,
				tracer.RPCMessageIDKey.String(routeName),
				tracer.RPCMessageCompressedSizeKey.Int(buf.Len()),
				tracer.InstanceKind.String(c.cli.Opts.InsKind.String()),
				tracer.InstanceID.String(c.cli.Opts.InsID),
				tracer.ServerIP.String(c.cli.Opts.Addr),
			}, attr...)...))

	return myCtx, span, func() { span.End() }, buf
}

// Trigger 触发事件
func (c *Client) Trigger(ctx context.Context, event cluster.Event, cid, uid int64) error {
	ctx, span, end, buf := c.traceBuffer(ctx, route.Trigger, 0, protocol.EncodeTriggerReq(event, cid, uid))
	defer end()

	if span != nil {
		span.SetName(fmt.Sprintf("node.client.%s.%s", route.Name[route.Trigger], event.String()))
	}

	return c.cli.Send(ctx, buf)
}

// Deliver 投递消息
func (c *Client) Deliver(ctx context.Context, cid, uid int64, message []byte) error {
	ctx, _, end, buf := c.traceBuffer(ctx, route.Deliver, 0, protocol.EncodeDeliverReq(cid, uid, message))
	defer end()
	return c.cli.Send(ctx, buf, cid)
}

// GetState 获取状态
func (c *Client) GetState(ctx context.Context) (cluster.State, error) {
	seq := c.doGenSequence()

	ctx, _, end, buf := c.traceBuffer(ctx, route.GetState, seq, nil)
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

	ctx, _, end, buf := c.traceBuffer(ctx, route.SetState, seq, protocol.EncodeSetStateReq(state))
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
