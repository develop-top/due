package protocol

import (
	"context"
	"encoding/binary"
	"github.com/develop-top/due/v2/core/buffer"
	"github.com/develop-top/due/v2/errors"
	"go.opentelemetry.io/otel/trace"
	"io"
)

const defaultTraceCtxBytes = 25 // 链路追踪上下文字节数

// 携带链路追踪数据
// 结构：消息长度4B + 链路追踪上下文25B + 消息
//TODO 原消息中前4字节数据冗余，需要去掉,所有编解码需要优化，在原有打好的包中修改数据要拷贝一次数据，希望不拷贝

// MarshalSpanContext 序列化 SpanContext -> []byte
func MarshalSpanContext(sc trace.SpanContext) []byte {
	buf := buffer.NewNocopyBuffer()
	traceID := sc.TraceID()
	spanID := sc.SpanID()
	buf.Mount(traceID[:])
	buf.Mount(spanID[:])
	buf.Mount([]byte{byte(sc.TraceFlags())})
	return buf.Bytes()
}

// UnmarshalSpanContext 反序列化 []byte -> SpanContext
func UnmarshalSpanContext(data []byte) trace.SpanContext {
	if len(data) < defaultTraceCtxBytes {
		return trace.SpanContext{} // 或报错
	}
	var traceID [16]byte
	var spanID [8]byte
	copy(traceID[:], data[0:16])
	copy(spanID[:], data[16:24])
	traceFlags := trace.TraceFlags(data[24])

	return trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: traceFlags,
		Remote:     true, // 这是接收到的 span
	})
}

// EncodeTraceMessage 序列化
func EncodeTraceMessage(data buffer.Buffer, traceCtx []byte) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(defaultSizeBytes)
	writer.WriteUint32s(binary.BigEndian, uint32(len(traceCtx)+data.Len()))
	buf.Mount(traceCtx)
	buf.Mount(data)
	return buf
}

func EncodeTraceBuffer(ctx context.Context, data buffer.Buffer) buffer.Buffer {
	traceCtx := MarshalSpanContext(trace.SpanContextFromContext(ctx))
	return EncodeTraceMessage(data, traceCtx)
}

// ReadTraceMessage 读取消息
func ReadTraceMessage(reader io.Reader) (isHeartbeat bool, route uint8, seq uint64, data, traceCtx []byte, err error) {
	buf := sizePool.Get().([]byte)

	if _, err = io.ReadFull(reader, buf); err != nil {
		sizePool.Put(buf)
		return
	}

	size := binary.BigEndian.Uint32(buf)

	if size == 0 {
		sizePool.Put(buf)
		err = errors.ErrInvalidMessage
		return
	}

	data = make([]byte, defaultSizeBytes+size)
	copy(data[:defaultSizeBytes], buf)

	sizePool.Put(buf)

	if _, err = io.ReadFull(reader, data[defaultSizeBytes:]); err != nil {
		return
	}

	traceCtx = data[defaultSizeBytes : defaultSizeBytes+defaultTraceCtxBytes]

	header := data[defaultSizeBytes+defaultTraceCtxBytes+defaultSizeBytes : defaultSizeBytes+defaultTraceCtxBytes+defaultSizeBytes+defaultHeaderBytes][0]

	isHeartbeat = header&heartbeatBit == heartbeatBit

	if isHeartbeat {
		return
	}

	route = data[defaultSizeBytes+defaultTraceCtxBytes+defaultSizeBytes+defaultHeaderBytes : defaultSizeBytes+defaultTraceCtxBytes+defaultSizeBytes+defaultHeaderBytes+defaultRouteBytes][0]

	seq = binary.BigEndian.Uint64(data[defaultSizeBytes+defaultTraceCtxBytes+defaultSizeBytes+defaultHeaderBytes+defaultRouteBytes : defaultSizeBytes+defaultTraceCtxBytes+defaultSizeBytes+defaultHeaderBytes+defaultRouteBytes+8])

	data = data[defaultSizeBytes+defaultTraceCtxBytes:]

	return
}
