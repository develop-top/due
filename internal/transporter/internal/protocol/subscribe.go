package protocol

import (
	"encoding/binary"
	"io"

	"github.com/develop-top/due/v2/core/buffer"
	"github.com/develop-top/due/v2/errors"
	"github.com/develop-top/due/v2/session"
)

const (
	subscribeReqBytes = b8 + b16
	subscribeResBytes = defaultCodeBytes
)

// EncodeSubscribeReq 编码订阅频道请求（单次最多订阅65535个对象）
// 协议：session kind + count + targets + channel
func EncodeSubscribeReq(kind session.Kind, targets []int64, channel string) buffer.Buffer {
	size := subscribeReqBytes + len(targets)*8 + len([]byte(channel))
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(size)
	writer.WriteUint8s(uint8(kind))
	writer.WriteUint16s(binary.BigEndian, uint16(len(targets)))
	writer.WriteInt64s(binary.BigEndian, targets...)
	writer.WriteString(channel)

	return buf
}

// DecodeSubscribeReq 解码订阅频道请求
// 协议：size + header + route + seq + [trace] + session kind + count + targets + channel
func DecodeSubscribeReq(data []byte) (seq uint64, kind session.Kind, targets []int64, channel string, err error) {
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

	channel = string(data[uint16(index)+subscribeReqBytes+8*count:])

	return
}

// EncodeSubscribeRes 编码订阅频道响应
// 协议：code
func EncodeSubscribeRes(code uint16) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(subscribeResBytes)
	writer.WriteUint16s(binary.BigEndian, code)

	return buf
}

// DecodeSubscribeRes 解码订阅频道响应
// 协议：size + header + route + seq + [trace] + code
func DecodeSubscribeRes(data []byte) (code uint16, err error) {
	if len(data) < SizeHeadRouteSeqBytes+subscribeResBytes {
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
