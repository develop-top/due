package server

import (
	"context"
	"github.com/develop-top/due/v2/cluster"
	"github.com/develop-top/due/v2/core/buffer"
	"github.com/develop-top/due/v2/errors"
	"github.com/develop-top/due/v2/internal/transporter/internal/def"
	"github.com/develop-top/due/v2/internal/transporter/internal/protocol"
	"github.com/develop-top/due/v2/log"
	"github.com/develop-top/due/v2/tracer"
	"github.com/develop-top/due/v2/utils/xtime"
	"github.com/develop-top/due/v2/utils/xtrace"
	"go.opentelemetry.io/otel/trace"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Conn struct {
	ctx               context.Context    // 上下文
	cancel            context.CancelFunc // 取消函数
	server            *Server            // 连接管理
	rw                sync.RWMutex       // 锁
	conn              net.Conn           // TCP源连接
	state             int32              // 连接状态
	chData            chan chData        // 消息处理通道
	lastHeartbeatTime int64              // 上次心跳时间
	InsKind           cluster.Kind       // 集群类型
	InsID             string             // 集群ID
}

func newConn(server *Server, conn net.Conn) *Conn {
	c := &Conn{}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.conn = conn
	c.server = server
	c.state = def.ConnOpened
	c.chData = make(chan chData, 10240)
	c.lastHeartbeatTime = xtime.Now().Unix()

	go c.read()

	go c.process()

	return c
}

// Send 发送消息
func (c *Conn) Send(ctx context.Context, buf buffer.Buffer) (err error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	if err = c.checkState(); err != nil {
		return err
	}

	// 携带链路追踪信息
	ctx, span := xtrace.StartRPCServerSpan(ctx, "internal.RPCServer", tracer.RPCMessageTypeSent)
	defer span.End()
	buf = protocol.EncodeTraceBuffer(ctx, buf)
	buf.Range(func(node *buffer.NocopyNode) bool {
		if _, err = c.conn.Write(node.Bytes()); err != nil {
			return false
		}
		return true
	})

	buf.Release()

	return
}

// 检测连接状态
func (c *Conn) checkState() error {
	if atomic.LoadInt32(&c.state) == def.ConnClosed {
		return errors.ErrConnectionClosed
	} else {
		return nil
	}
}

// 关闭连接
func (c *Conn) close(isNeedRecycle ...bool) error {
	if !atomic.CompareAndSwapInt32(&c.state, def.ConnOpened, def.ConnClosed) {
		return errors.ErrConnectionClosed
	}

	c.rw.Lock()
	defer c.rw.Unlock()

	c.cancel()

	close(c.chData)

	if len(isNeedRecycle) > 0 && isNeedRecycle[0] {
		c.server.recycle(c.conn)
	}

	return c.conn.Close()
}

// 读取消息
func (c *Conn) read() {
	conn := c.conn

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			isHeartbeat, route, _, data, traceCtx, err := protocol.ReadTraceMessage(conn)
			if err != nil {
				_ = c.close(true)
				return
			}

			c.rw.RLock()

			if atomic.LoadInt32(&c.state) == def.ConnClosed {
				c.rw.RUnlock()
				return
			}

			c.chData <- chData{
				isHeartbeat: isHeartbeat,
				route:       route,
				data:        data,
				trace:       traceCtx,
			}

			c.rw.RUnlock()
		}
	}
}

// 处理数据
func (c *Conn) process() {
	ticker := time.NewTicker(def.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			deadline := xtime.Now().Add(-2 * def.HeartbeatInterval).Unix()
			if atomic.LoadInt64(&c.lastHeartbeatTime) < deadline {
				_ = c.close(true)
				return
			}
		case ch, ok := <-c.chData:
			if !ok {
				return
			}

			atomic.StoreInt64(&c.lastHeartbeatTime, xtime.Now().Unix())

			if ch.isHeartbeat {
				c.heartbeat(ch)
			} else {
				handler, ok := c.server.handlers[ch.route]
				if !ok {
					continue
				}

				// 携带链路追踪信息
				ctx := trace.ContextWithRemoteSpanContext(context.Background(), protocol.UnmarshalSpanContext(ch.trace))
				ctx, span := xtrace.StartRPCServerSpan(ctx, "internal.RPCServer", tracer.RPCMessageTypeReceived)
				if err := handler(ctx, c, ch.data); err != nil && !errors.Is(err, errors.ErrNotFoundUserLocation) {
					log.Warnf("process route %d message failed: %v", ch.route, err)
				}
				span.End()
			}
		}
	}
}

// 响应心跳消息
func (c *Conn) heartbeat(ch chData) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	// 携带链路追踪信息
	ctx := trace.ContextWithRemoteSpanContext(context.Background(), protocol.UnmarshalSpanContext(ch.trace))
	myCtx, span := xtrace.StartRPCServerSpan(ctx, "internal.RPCServer.Heartbeat", tracer.RPCMessageTypeSent)
	defer span.End()
	buf := protocol.EncodeTraceBuffer(myCtx, buffer.NewNocopyBuffer(protocol.Heartbeat()))
	defer buf.Release()
	if _, err := c.conn.Write(buf.Bytes()); err != nil {
		log.Warnf("write heartbeat message error: %v", err)
	}
}
