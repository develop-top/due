package protocol

import (
	"encoding/binary"
	"github.com/develop-top/due/v2/core/buffer"
	"github.com/develop-top/due/v2/errors"
	"go.opentelemetry.io/otel/trace"
	"io"
)

const (
	SizeHeadRouteSeqBytes = defaultSizeBytes + defaultHeaderBytes + defaultRouteBytes + defaultSeqBytes
)

// EncodeBuffer 组装消息包
// size + head + route + seq + [trace] + [other]
func EncodeBuffer(head, route uint8, seq uint64, trace []byte, other buffer.Buffer) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(SizeHeadRouteSeqBytes)
	size := uint32(SizeHeadRouteSeqBytes - defaultSizeBytes + len(trace))
	if other != nil {
		size += uint32(other.Len())
	}
	writer.WriteUint32s(binary.BigEndian, size)
	if len(trace) > 0 {
		head |= TraceBit
	}
	writer.WriteUint8s(head)
	writer.WriteUint8s(route)
	writer.WriteUint64s(binary.BigEndian, seq)
	if len(trace) > 0 {
		buf.Mount(trace)
	}
	if other != nil && other.Len() > 0 {
		buf.Mount(other)
	}
	return buf
}

func DecodeHeader(reader *buffer.Reader) (header uint8, err error) {
	if _, err = reader.Seek(defaultSizeBytes, io.SeekStart); err != nil {
		return
	}

	header, err = reader.ReadUint8()
	if err != nil {
		return
	}

	return header, nil
}

func DecodeSeq(reader *buffer.Reader) (seq uint64, err error) {
	if _, err = reader.Seek(defaultSizeBytes+defaultHeaderBytes+defaultRouteBytes, io.SeekStart); err != nil {
		return
	}

	seq, err = reader.ReadUint64(binary.BigEndian)
	if err != nil {
		return
	}

	return seq, nil
}

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
	if len(data) < defaultTraceBytes {
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

// ReadTraceMessage 读取消息
// size + head + route + seq + [trace] + [other]
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

	header := data[defaultSizeBytes : defaultSizeBytes+defaultHeaderBytes][0]

	isHeartbeat = header&HeartbeatBit == HeartbeatBit

	if isHeartbeat {
		return
	}

	route = data[defaultSizeBytes+defaultHeaderBytes : defaultSizeBytes+defaultHeaderBytes+defaultRouteBytes][0]

	seq = binary.BigEndian.Uint64(data[defaultSizeBytes+defaultHeaderBytes+defaultRouteBytes : defaultSizeBytes+defaultHeaderBytes+defaultRouteBytes+defaultSeqBytes])

	hasTrace := header&TraceBit == TraceBit

	if hasTrace {
		traceCtx = data[SizeHeadRouteSeqBytes : SizeHeadRouteSeqBytes+defaultTraceBytes]
	}

	return
}
