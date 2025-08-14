package protocol

import (
	"encoding/binary"
	"io"

	"github.com/develop-top/due/v2/core/buffer"
	"github.com/develop-top/due/v2/errors"
	"github.com/develop-top/due/v2/session"
)

const (
	unsubscribeReqBytes = b8 + b16
	unsubscribeResBytes = defaultCodeBytes
)

// EncodeUnsubscribeReq 编码取消订阅频道请求（单次最多取消订阅65535个对象）
// 协议：session kind + count + targets + channel
func EncodeUnsubscribeReq(kind session.Kind, targets []int64, channel string) buffer.Buffer {
	size := unsubscribeReqBytes + len(targets)*8 + len([]byte(channel))
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(size)
	writer.WriteUint8s(uint8(kind))
	writer.WriteUint16s(binary.BigEndian, uint16(len(targets)))
	writer.WriteInt64s(binary.BigEndian, targets...)
	writer.WriteString(channel)

	return buf
}

// DecodeUnsubscribeReq 解码取消订阅频道请求
// 协议：size + header + route + seq + [trace] + session kind + count + targets + channel
func DecodeUnsubscribeReq(data []byte) (seq uint64, kind session.Kind, targets []int64, channel string, err error) {
	reader := buffer.NewReader(data)

	seq, err = DecodeSeq(reader)
	if err != nil {
		return
	}

	var header uint8
	if header, err = DecodeHeader(reader); err != nil {
		return
	}
	index := SizeHeadRouteSeqBytes
	if header&TraceBit == TraceBit {
		index += defaultTraceBytes
	}

	if _, err = reader.Seek(int64(index), io.SeekStart); err != nil {
		return
	}

	var k uint8
	if k, err = reader.ReadUint8(); err != nil {
		return
	} else {
		kind = session.Kind(k)
	}

	count, err := reader.ReadUint16(binary.BigEndian)
	if err != nil {
		return
	}

	if targets, err = reader.ReadInt64s(binary.BigEndian, int(count)); err != nil {
		return
	}

	channel = string(data[uint16(index)+unsubscribeReqBytes+8*count:])

	return
}

// EncodeUnsubscribeRes 编码取消订阅频道响应
// 协议：code
func EncodeUnsubscribeRes(code uint16) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(unsubscribeResBytes)
	writer.WriteUint16s(binary.BigEndian, code)

	return buf
}

// DecodeUnsubscribeRes 解码取消订阅频道响应
// 协议：size + header + route + seq + [trace] + code
func DecodeUnsubscribeRes(data []byte) (code uint16, err error) {
	if len(data) < SizeHeadRouteSeqBytes+unsubscribeResBytes {
		err = errors.ErrInvalidMessage
		return
	}

	reader := buffer.NewReader(data)

	if _, err = reader.Seek(-defaultCodeBytes, io.SeekEnd); err != nil {
		return
	}

	if code, err = reader.ReadUint16(binary.BigEndian); err != nil {
		return
	}

	return
}
